package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommandMissingDependenciesSuppressUsage(t *testing.T) {
	for _, explicitOwner := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit owner=%t", explicitOwner), func(t *testing.T) {
			dir := t.TempDir()
			if !explicitOwner {
				doctorExecutable(t, dir, "gh")
			}
			t.Setenv("PATH", dir)
			t.Setenv("GHAPP_OWNER", "")
			cmd, err := newRootCmd(rootDeps{})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			args := []string{"--", "pr", "list"}
			missing := "git"
			if explicitOwner {
				args = append([]string{"--owner", "owner"}, args...)
				missing = "gh"
			}
			cmd.SetArgs(args)
			require.ErrorContains(t, cmd.ExecuteContext(t.Context()), missing)
			assert.Contains(t, output.String(), "Error:")
			assert.NotContains(t, output.String(), "Usage:")
		})
	}
}

func TestRootCommandRuntimeErrorSuppressesUsage(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("owner resolution failed")
	cmd, err := newRootCmd(rootDeps{
		ownerResolver: &OwnerResolverMock{
			OwnerFunc: func(context.Context, string) (string, error) {
				return "", wantErr
			},
		},
	})
	require.NoError(t, err)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"--", "pr", "list"})

	require.ErrorIs(t, cmd.ExecuteContext(t.Context()), wantErr)
	assert.Contains(t, output.String(), "Error: owner resolution failed")
	assert.NotContains(t, output.String(), "Usage:")
}

func TestRootCommandTokenAndExecutorErrorsSuppressUsage(t *testing.T) {
	t.Parallel()

	exitErr := exec.CommandContext(t.Context(), "sh", "-c", "exit 42").Run()
	var wantExitErr *exec.ExitError
	require.ErrorAs(t, exitErr, &wantExitErr)
	for _, stage := range []string{"token", "executor", "executor exit"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			wantErr := errors.New(stage + " failed")
			if stage == "executor exit" {
				wantErr = exitErr
			}
			cmd, err := newRootCmd(rootDeps{
				ownerResolver: &OwnerResolverMock{
					OwnerFunc: func(context.Context, string) (string, error) { return "owner", nil },
				},
				tokenGenerator: &TokenGeneratorMock{
					NewFunc: func(context.Context, string, string, string) (string, error) {
						if stage == "token" {
							return "", wantErr
						}
						return "token", nil
					},
				},
				ghcliExecutor: &GHCliExecutorMock{
					ExecFunc: func(context.Context, string, []string) error { return wantErr },
				},
			})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs([]string{"--", "pr", "list"})

			actualErr := cmd.ExecuteContext(t.Context())
			require.ErrorIs(t, actualErr, wantErr)
			assert.Contains(t, output.String(), "Error: "+wantErr.Error())
			assert.NotContains(t, output.String(), "Usage:")
			if stage == "executor exit" {
				var actualExitErr *exec.ExitError
				require.ErrorAs(t, actualErr, &actualExitErr)
				assert.Equal(t, 42, actualExitErr.ExitCode())
			}
		})
	}
}

func TestRootCommandUsagePolicy(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		args      []string
		wantError string
		wantText  string
	}{
		{name: "missing separator", args: []string{}, wantError: "Specify that args for gh after", wantText: "Usage:"},
		{name: "argument before separator", args: []string{"pr", "--", "list"}, wantError: "Do not pass position args before", wantText: "Usage:"},
		{name: "unknown flag", args: []string{"--unknown"}, wantError: "unknown flag", wantText: "Usage:"},
		{name: "help", args: []string{"--help"}, wantText: "Usage:"},
		{name: "skills", args: []string{"skills"}, wantText: "# GitHub App authentication"},
		{name: "skills help", args: []string{"skills", "--help"}, wantText: "Usage:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Nil dependencies ensure help, subcommands, and invalid args never run the root.
			cmd, err := newRootCmd(rootDeps{})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tc.args)
			err = cmd.ExecuteContext(t.Context())
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				assert.Contains(t, output.String(), "Error:")
			} else {
				require.NoError(t, err)
			}
			assert.Contains(t, output.String(), tc.wantText)
			assert.False(t, cmd.SilenceUsage)
		})
	}
}

func TestRootCommand(t *testing.T) {
	dir := t.TempDir()
	doctorExecutable(t, dir, "gh")
	t.Setenv("PATH", dir)

	expectedOwner := "foobar"
	expectedClientID := "this.is.client.id"
	expectedPrivateKeyPath := "/path/to/pk/file"
	expectedToken := "foo.bar.token"
	expectedArgs := []string{"pr", "list"}
	cmd, err := newRootCmd(rootDeps{
		ownerResolver: &OwnerResolverMock{
			OwnerFunc: func(_ context.Context, actualExplicitOwner string) (string, error) {
				assert.Equal(t, expectedOwner, actualExplicitOwner)
				return actualExplicitOwner, nil
			},
		},
		tokenGenerator: &TokenGeneratorMock{
			NewFunc: func(_ context.Context, actualOwner string, actualClientID string, actualPrivateKeyPath string) (string, error) {
				assert.Equal(t, expectedOwner, actualOwner)
				assert.Equal(t, expectedClientID, actualClientID)
				assert.Equal(t, expectedPrivateKeyPath, actualPrivateKeyPath)
				return expectedToken, nil
			},
		},
		ghcliExecutor: &GHCliExecutorMock{
			ExecFunc: func(_ context.Context, actualToken string, actualArgs []string) error {
				assert.Equal(t, expectedToken, actualToken)
				assert.Equal(t, expectedArgs, actualArgs)
				return nil
			},
		},
	})
	require.NoError(t, err)

	cmd.SetArgs(append(
		[]string{
			"--owner", expectedOwner,
			"--client-id", expectedClientID,
			"--private-key-path", expectedPrivateKeyPath,
			"--",
		},
		expectedArgs...,
	))

	require.NoError(t, cmd.ExecuteContext(t.Context()))
}
