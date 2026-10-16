package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func doctorExecutable(t *testing.T, dir, name string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(dir, name)
	// LookPath checks executability, not executable file contents.
	require.NoError(t, os.WriteFile(path, []byte("fixture\n"), 0o755))
	return path
}

func TestDoctorMissingDependencies(t *testing.T) {
	for _, available := range []string{"gh", "git", ""} {
		t.Run("available="+available, func(t *testing.T) {
			dir := t.TempDir()
			var availablePath string
			if available != "" {
				availablePath = doctorExecutable(t, dir, available)
			}
			t.Setenv("PATH", dir)
			if runtime.GOOS == "windows" {
				t.Setenv("PATHEXT", ".EXE")
			}
			cmd, err := newRootCmd(rootDeps{})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs([]string{"doctor"})

			require.Error(t, cmd.ExecuteContext(t.Context()))
			for _, name := range []string{"gh", "git"} {
				if name == available {
					require.Contains(t, output.String(), name+": "+availablePath)
				} else {
					require.Contains(t, output.String(), name+": not found")
				}
			}
		})
	}
}

func TestDoctorRejectsExtraArgs(t *testing.T) {
	cmd, err := newRootCmd(rootDeps{})
	require.NoError(t, err)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"doctor", "extra"})

	require.ErrorContains(t, cmd.ExecuteContext(t.Context()), "unknown command")
	require.NotContains(t, output.String(), "gh:")
	require.NotContains(t, output.String(), "git:")
}

func TestDoctorCommand(t *testing.T) {
	dir := t.TempDir()
	ghPath := doctorExecutable(t, dir, "gh")
	gitPath := doctorExecutable(t, dir, "git")
	t.Setenv("PATH", dir)
	if runtime.GOOS == "windows" {
		t.Setenv("PATHEXT", ".EXE")
	}

	cmd, err := newRootCmd(rootDeps{})
	require.NoError(t, err)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"doctor"})

	require.NoError(t, cmd.ExecuteContext(t.Context()))
	require.Contains(t, output.String(), "gh: "+ghPath)
	require.Contains(t, output.String(), "git: "+gitPath)
}
