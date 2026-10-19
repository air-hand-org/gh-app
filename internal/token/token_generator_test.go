package token

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratorRejectsUnsafeConfigurationOffline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sensitive-path")
	generator := NewInstallationTokenGenerator()
	for _, tc := range []struct {
		clientID, path string
		expected       error
	}{
		{"", path, ErrNotConfigured},
		{"sensitive-client", path, ErrKeyNotExist},
	} {
		value, err := generator.New(t.Context(), "owner", tc.clientID, tc.path)
		require.Empty(t, value)
		require.ErrorIs(t, err, tc.expected)
		require.Nil(t, errors.Unwrap(err))
		require.NotContains(t, fmt.Sprintf("%+v", err), path)
	}
	require.NoError(t, os.WriteFile(path, []byte("sensitive-key"), 0o600))
	value, err := generator.New(t.Context(), "owner", "client", path)
	require.Empty(t, value)
	require.ErrorIs(t, err, ErrInvalidKey)
	require.NotContains(t, err.Error(), "sensitive-key")
}

func TestInspectInstallationID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		orgID     int64
		orgErr    error
		userID    int64
		userErr   error
		expected  int64
		expectErr bool
	}{
		{
			name:      "org only",
			orgID:     100,
			orgErr:    nil,
			userID:    200,
			userErr:   errors.New("failed to get user id"),
			expected:  100,
			expectErr: false,
		},
		{
			name:      "user only",
			orgID:     100,
			orgErr:    errors.New("failed to get org id"),
			userID:    200,
			userErr:   nil,
			expected:  200,
			expectErr: false,
		},
		{
			name:      "org and user fail",
			orgID:     100,
			orgErr:    errors.New("failed to get org id"),
			userID:    200,
			userErr:   errors.New("failed to get user id"),
			expected:  0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			getter := &InstallationIDGetterMock{
				OrganizationInstallationIDFunc: func(_ context.Context, _ string) (int64, error) {
					return tt.orgID, tt.orgErr
				},
				UserInstallationIDFunc: func(_ context.Context, _ string) (int64, error) {
					return tt.userID, tt.userErr
				},
			}

			actual, err := inspectInstallationID(t.Context(), tt.name, getter)
			if tt.expectErr {
				require.True(t, err != nil)
			} else {
				require.Equal(t, tt.expected, actual)
			}
		})
	}
}
