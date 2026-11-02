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

func TestIdentityCommandsRejectInvalidArguments(t *testing.T) {
	for _, args := range [][]string{{"identity", "extra"}, {"git-config", "extra"}, {"git-config", "--local", "--global"}, {"git-config", "--local=false", "--global"}} {
		t.Run(args[0]+args[len(args)-1], func(t *testing.T) {
			cmd, err := newRootCmd(rootDeps{})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(args)
			require.Error(t, cmd.ExecuteContext(t.Context()))
			require.Contains(t, output.String(), "Usage:")
		})
	}
}

func TestIdentityCommandPlainTextAndErrors(t *testing.T) {
	for _, fails := range []bool{false, true} {
		t.Run(map[bool]string{true: "error", false: "plain text"}[fails], func(t *testing.T) {
			wantErr := errors.New("authentication failed")
			cmd, err := newRootCmd(rootDeps{identityResolver: identityResolverFunc(func(context.Context, string, string) (token.Identity, error) {
				if fails {
					return token.Identity{}, wantErr
				}
				return token.Identity{Name: "app[bot]", Email: "456+app[bot]@users.noreply.github.com"}, nil
			})})
			require.NoError(t, err)
			var output, stderr bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&stderr)
			cmd.SetArgs([]string{"identity"})
			err = cmd.ExecuteContext(t.Context())
			if fails {
				require.ErrorIs(t, err, wantErr)
				require.Empty(t, output.String())
				require.NotContains(t, stderr.String(), "Usage:")
			} else {
				require.NoError(t, err)
				require.Equal(t, "app[bot] <456+app[bot]@users.noreply.github.com>\n", output.String())
			}
		})
	}
}

func TestGitConfigFailuresDoNotWrite(t *testing.T) {
	for _, stage := range []string{"missing git", "outside repository", "authentication", "locked config"} {
		t.Run(stage, func(t *testing.T) {
			home := t.TempDir()
			global := filepath.Join(home, "global")
			t.Setenv("HOME", home)
			t.Setenv("XDG_CONFIG_HOME", home)
			t.Setenv("GIT_CONFIG_GLOBAL", global)
			t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
			t.Setenv("GIT_CONFIG_COUNT", "0")
			t.Chdir(t.TempDir())
			if stage != "outside repository" {
				require.NoError(t, exec.CommandContext(t.Context(), "git", "init", "--quiet").Run())
			}
			if stage == "missing git" {
				t.Setenv("PATH", t.TempDir())
			}
			if stage == "locked config" {
				require.NoError(t, os.WriteFile(filepath.Join(".git", "config.lock"), nil, 0o600))
			}
			called := false
			cmd, err := newRootCmd(rootDeps{identityResolver: identityResolverFunc(func(context.Context, string, string) (token.Identity, error) {
				called = true
				if stage == "authentication" {
					return token.Identity{}, errors.New("authentication failed")
				}
				return token.Identity{Name: "app[bot]", Email: "456+app[bot]@users.noreply.github.com"}, nil
			})})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs([]string{"git-config"})
			require.Error(t, cmd.ExecuteContext(t.Context()))
			require.Equal(t, stage == "authentication" || stage == "locked config", called)
			require.NotContains(t, output.String(), "Usage:")
			_, err = os.Stat(global)
			require.True(t, os.IsNotExist(err))
			if stage != "outside repository" {
				contents, err := os.ReadFile(filepath.Join(".git", "config"))
				require.NoError(t, err)
				require.NotContains(t, string(contents), "app[bot]")
			}
		})
	}
}

func TestGitConfigGlobalOutsideRepository(t *testing.T) {
	home := t.TempDir()
	global := filepath.Join(home, "global")
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_COUNT", "0")
	t.Chdir(t.TempDir())
	cmd, err := newRootCmd(rootDeps{identityResolver: identityResolverFunc(func(context.Context, string, string) (token.Identity, error) {
		return token.Identity{Name: "app[bot]", Email: "456+app[bot]@users.noreply.github.com"}, nil
	})})
	require.NoError(t, err)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"git-config", "--global"})
	require.NoError(t, cmd.ExecuteContext(t.Context()))
	contents, err := os.ReadFile(global)
	require.NoError(t, err)
	require.Contains(t, string(contents), "app[bot]")
	require.Contains(t, string(contents), "456+app[bot]@users.noreply.github.com")
}
