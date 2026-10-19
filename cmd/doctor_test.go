package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func doctorExecutable(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	// LookPath checks executability, not executable file contents.
	require.NoError(t, os.WriteFile(path, []byte("fixture\n"), 0o755))
	return path
}

func doctorKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "private-key.pem")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), 0o600))
	return path
}

func TestDoctorFlagPrecedence(t *testing.T) {
	for _, args := range [][]string{
		{"doctor", "--client-id", "flag-client", "--private-key-path"},
		{"--client-id", "flag-client", "--private-key-path"},
	} {
		t.Run(args[0], func(t *testing.T) {
			dir := t.TempDir()
			doctorExecutable(t, dir, "gh")
			doctorExecutable(t, dir, "git")
			t.Setenv("PATH", dir)
			t.Setenv("GHAPP_CLIENT_ID", "env-client")
			t.Setenv("GHAPP_PRIVATE_KEY_PATH", "/nonexistent-env-path")
			path := doctorKey(t)
			commandArgs := append(append([]string{}, args...), path)
			if args[0] != "doctor" {
				commandArgs = append(commandArgs, "doctor")
			}
			cmd, err := newRootCmd(rootDeps{})
			require.NoError(t, err)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(commandArgs)
			require.NoError(t, cmd.ExecuteContext(t.Context()))
			require.Contains(t, output.String(), "client-id: OK (configured)")
			require.Contains(t, output.String(), "private-key: OK (supported RSA private key)")
			for _, secret := range []string{"flag-client", "env-client", "/nonexistent-env-path", path} {
				require.NotContains(t, output.String(), secret)
			}
		})
	}
}

func TestDoctorExplicitEmptyClientID(t *testing.T) {
	dir := t.TempDir()
	doctorExecutable(t, dir, "gh")
	doctorExecutable(t, dir, "git")
	t.Setenv("PATH", dir)
	t.Setenv("GHAPP_CLIENT_ID", "environment-client")
	t.Setenv("GHAPP_PRIVATE_KEY_PATH", doctorKey(t))
	cmd, err := newRootCmd(rootDeps{})
	require.NoError(t, err)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"doctor", "--client-id="})
	require.ErrorContains(t, cmd.ExecuteContext(t.Context()), "local configuration checks failed")
	require.Contains(t, output.String(), "client-id: FAIL (not configured)")
	require.Contains(t, output.String(), "private-key: OK (supported RSA private key)")
	require.NotContains(t, output.String(), "environment-client")
}

func TestDoctorMissingConfiguration(t *testing.T) {
	dir := t.TempDir()
	doctorExecutable(t, dir, "gh")
	doctorExecutable(t, dir, "git")
	t.Setenv("PATH", dir)
	t.Setenv("GHAPP_CLIENT_ID", "")
	t.Setenv("GHAPP_PRIVATE_KEY_PATH", "")
	cmd, err := newRootCmd(rootDeps{})
	require.NoError(t, err)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"doctor"})
	require.ErrorContains(t, cmd.ExecuteContext(t.Context()), "local configuration checks failed")
	require.Contains(t, output.String(), "client-id: FAIL (not configured)")
	require.Contains(t, output.String(), "private-key-path: FAIL (not configured)")
	require.Contains(t, output.String(), "private-key: SKIP (file unavailable)")
	require.NotContains(t, output.String(), "Usage:")
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
			t.Setenv("GHAPP_CLIENT_ID", "")
			t.Setenv("GHAPP_PRIVATE_KEY_PATH", "")
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
	t.Setenv("GHAPP_CLIENT_ID", "test-client")
	t.Setenv("GHAPP_PRIVATE_KEY_PATH", doctorKey(t))

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

func TestDoctorInvalidKeyOutput(t *testing.T) {
	dir := t.TempDir()
	doctorExecutable(t, dir, "gh")
	doctorExecutable(t, dir, "git")
	t.Setenv("PATH", dir)
	t.Setenv("GHAPP_CLIENT_ID", "sensitive-client")
	path := filepath.Join(t.TempDir(), "sensitive-path")
	require.NoError(t, os.WriteFile(path, []byte("sensitive-key-contents"), 0o600))
	t.Setenv("GHAPP_PRIVATE_KEY_PATH", path)
	cmd, err := newRootCmd(rootDeps{})
	require.NoError(t, err)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"doctor"})
	require.ErrorContains(t, cmd.ExecuteContext(t.Context()), "local configuration checks failed")
	require.Contains(t, output.String(), "client-id: OK (configured)")
	require.Contains(t, output.String(), "private-key-path: OK (readable regular file)")
	require.Contains(t, output.String(), "private-key: FAIL (invalid or unsupported private key)")
	for _, secret := range []string{path, "sensitive-client", "sensitive-key-contents", "Usage:"} {
		require.NotContains(t, output.String(), secret)
	}
}
