package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootDependencyChecks(t *testing.T) {
	for _, tt := range []struct {
		name      string
		available []string
		args      []string
		envOwner  string
		missing   string
	}{
		{name: "gh always required", available: []string{"git"}, args: []string{"--owner", "org", "--"}, missing: "gh"},
		{name: "git needed for detection", available: []string{"gh"}, args: []string{"--"}, missing: "git"},
		{name: "explicit owner skips git", available: []string{"gh"}, args: []string{"--owner", "org", "--"}},
		{name: "environment owner skips git", available: []string{"gh"}, args: []string{"--"}, envOwner: "org"},
		{name: "empty owner needs git", available: []string{"gh"}, args: []string{"--owner", "", "--"}, missing: "git"},
		{name: "detected owner", available: []string{"gh", "git"}, args: []string{"--"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range tt.available {
				doctorExecutable(t, dir, name)
			}
			t.Setenv("PATH", dir)
			t.Setenv("GHAPP_OWNER", tt.envOwner)
			var calls []string
			cmd, err := newRootCmd(rootDeps{
				ownerResolver: &OwnerResolverMock{OwnerFunc: func(_ context.Context, _ string) (string, error) { calls = append(calls, "owner"); return "org", nil }},
				tokenGenerator: &TokenGeneratorMock{NewFunc: func(_ context.Context, _, _, _ string) (string, error) {
					calls = append(calls, "token")
					return "token", nil
				}},
				ghcliExecutor: &GHCliExecutorMock{ExecFunc: func(_ context.Context, _ string, _ []string) error { calls = append(calls, "exec"); return nil }},
			})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)
			err = cmd.ExecuteContext(t.Context())
			if tt.missing != "" {
				require.ErrorContains(t, err, tt.missing)
				require.Empty(t, calls)
			} else {
				require.NoError(t, err)
				require.Equal(t, []string{"owner", "token", "exec"}, calls)
				require.Empty(t, output.String())
			}
		})
	}
}

func TestRootDependencyChecksDoNotAffectOtherCommands(t *testing.T) {
	for _, args := range [][]string{{"skills"}, {"--help"}, {"help"}, {"--version"}, {"invalid"}} {
		t.Run(args[0], func(t *testing.T) {
			t.Setenv("PATH", t.TempDir())
			cmd, err := newRootCmd(rootDeps{})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(args)
			err = cmd.ExecuteContext(t.Context())
			if args[0] == "invalid" {
				require.ErrorContains(t, err, "Specify that args")
			} else {
				require.NoError(t, err)
			}
			require.NotContains(t, output.String(), "executable file not found")
		})
	}
}
