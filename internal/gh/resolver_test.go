package gh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOwnerFromGitRemote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:      "https",
			input:     "https://github.com/owner/repo.git",
			expected:  "owner",
			expectErr: false,
		},
		{
			name:      "https without .git",
			input:     "https://github.com/owner/repo",
			expected:  "owner",
			expectErr: false,
		},
		{
			name:      "https but not github.com",
			input:     "https://jp.github.com/owner/repo",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "https but without repo",
			input:     "https://jp.github.com/owner/",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "https but empty owner",
			input:     "https://jp.github.com//repo",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "ssh",
			input:     "ssh://git@github.com/owner/repo.git",
			expected:  "owner",
			expectErr: false,
		},
		{
			name:      "scp",
			input:     "git@github.com:owner/repo.git",
			expected:  "owner",
			expectErr: false,
		},
		{
			name:      "scp but without repo",
			input:     "git@github.com:owner/",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "api url",
			input:     "https://api.github.com/repos/owner/repo/issues",
			expected:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseOwnerFromGitRemote(tt.input)
			require.Equal(t, tt.expectErr, err != nil)
			require.Equal(t, tt.expected, actual)
		})
	}
}
