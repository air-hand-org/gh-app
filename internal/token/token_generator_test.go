package token

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

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
