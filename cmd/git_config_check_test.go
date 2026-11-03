package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/air-hand-org/gh-app/internal/token"
	"github.com/stretchr/testify/require"
)

func TestGitConfigCheck(t *testing.T) {
	for _, tc := range []struct{ name, local, want string }{
		{"global match", "", ""},
		{"local match", "[user]\nname = app[bot]\n", ""},
		{"mismatch", "[user]\nname = Other\n", `Git user.name "Other" does not match authenticated App bot "app[bot]"`},
		{"unset", "", "Git user.name is unset"},
		{"empty", "[user]\nname =\n", `Git user.name "" does not match`},
		{"git error", "invalid config\n", "read effective Git user.name"},
		{"auth error", "", "authentication failed"},
		{"missing git", "", "git"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_CONFIG_HOME", home)
			global := filepath.Join(home, "global")
			t.Setenv("GIT_CONFIG_GLOBAL", global)
			t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
			t.Setenv("GIT_CONFIG_COUNT", "0")
			t.Chdir(t.TempDir())
			require.NoError(t, exec.CommandContext(t.Context(), "git", "init", "--quiet").Run())
			require.NoError(t, os.WriteFile(global, []byte("[user]\nname = app[bot]\nemail = unrelated@example.com\n"), 0o600))
			if tc.name == "unset" {
				require.NoError(t, os.WriteFile(global, nil, 0o600))
			}
			if tc.local != "" {
				require.NoError(t, os.WriteFile(".git/config", []byte(tc.local), 0o600))
			}
			if tc.name == "missing git" {
				t.Setenv("PATH", t.TempDir())
			}
			before, err := os.ReadFile(global)
			require.NoError(t, err)
			localBefore, err := os.ReadFile(".git/config")
			require.NoError(t, err)
			cmd, err := newRootCmd(rootDeps{identityResolver: identityResolverFunc(func(context.Context, string, string) (token.Identity, error) {
				if tc.name == "auth error" {
					return token.Identity{}, errors.New("authentication failed")
				}
				return token.Identity{Name: "app[bot]"}, nil
			})})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs([]string{"git-config", "check"})
			err = cmd.ExecuteContext(t.Context())
			if tc.want == "" {
				require.NoError(t, err)
				require.Contains(t, output.String(), "app[bot]")
			} else {
				require.ErrorContains(t, err, tc.want)
				require.NotContains(t, output.String(), "matches authenticated")
			}
			require.NotContains(t, output.String(), "Usage:")
			after, err := os.ReadFile(global)
			require.NoError(t, err)
			require.Equal(t, before, after)
			localAfter, err := os.ReadFile(".git/config")
			require.NoError(t, err)
			require.Equal(t, localBefore, localAfter)
		})
	}
}
