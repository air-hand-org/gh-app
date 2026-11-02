package cmd

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/air-hand-org/gh-app/internal/token"
	"github.com/stretchr/testify/require"
)

func TestGitConfigCommand(t *testing.T) {
	for _, scope := range []string{"default", "local", "global"} {
		t.Run(scope, func(t *testing.T) {
			home := t.TempDir()
			global := filepath.Join(home, "global")
			t.Setenv("HOME", home)
			t.Setenv("XDG_CONFIG_HOME", home)
			t.Setenv("GIT_CONFIG_GLOBAL", global)
			t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
			t.Setenv("GIT_CONFIG_COUNT", "0")
			repo := t.TempDir()
			t.Chdir(repo)
			require.NoError(t, exec.CommandContext(t.Context(), "git", "init", "--quiet").Run())
			require.NoError(t, os.WriteFile(global, []byte("[user]\n\tname = Original\n\temail = original@example.com\n"), 0o600))
			cmd, err := newRootCmd(rootDeps{identityResolver: identityResolverFunc(func(context.Context, string, string) (token.Identity, error) {
				return token.Identity{ID: 456, Login: "app[bot]", Name: "app[bot]", Email: "456+app[bot]@users.noreply.github.com"}, nil
			})})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			args := []string{"git-config"}
			readScope := "--local"
			if scope != "default" {
				args = append(args, "--"+scope)
			}
			if scope == "global" {
				readScope = "--global"
			}
			cmd.SetArgs(args)
			require.NoError(t, cmd.ExecuteContext(t.Context()))
			for key, want := range map[string]string{"user.name": "app[bot]", "user.email": "456+app[bot]@users.noreply.github.com"} {
				got, err := exec.CommandContext(t.Context(), "git", "config", readScope, "--get", key).Output()
				require.NoError(t, err)
				require.Equal(t, want+"\n", string(got))
			}
			if scope != "global" {
				contents, err := os.ReadFile(global)
				require.NoError(t, err)
				require.Contains(t, string(contents), "Original")
				require.NotContains(t, string(contents), "app[bot]")
			} else {
				out, err := exec.CommandContext(t.Context(), "git", "config", "--local", "--get", "user.name").Output()
				require.Error(t, err)
				require.Empty(t, out)
			}
		})
	}
}
