package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/air-hand-org/gh-app/internal/skills"
	"github.com/stretchr/testify/require"
)

func runSkills(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd, err := newRootCmd(rootDeps{})
	require.NoError(t, err)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(append([]string{"skills"}, args...))
	err = cmd.ExecuteContext(t.Context())
	return out.String(), err
}

func TestSkillsInstallRequiresDirectory(t *testing.T) {
	for _, args := range [][]string{{"install"}, {"install", "--skills-dir", ""}} {
		_, err := runSkills(t, args...)
		require.ErrorContains(t, err, "skills-dir")
	}
}

func TestSkillsInstallValidatesBeforeRun(t *testing.T) {
	cmd, err := newRootCmd(rootDeps{})
	require.NoError(t, err)
	install, _, err := cmd.Find([]string{"skills", "install"})
	require.NoError(t, err)
	require.NotNil(t, install.PreRunE)
	require.ErrorContains(t, install.PreRunE(install, nil), "--skills-dir must not be empty")
	require.False(t, install.SilenceUsage)
}

func TestSkillsInstallInvalidInputDoesNotWrite(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		message string
	}{
		{name: "extra argument", args: []string{"--skills-dir", "skills", "extra"}, message: "unknown command"},
		{name: "empty directory", args: []string{"--skills-dir", "", "--force"}, message: "--skills-dir must not be empty"},
		{name: "missing directory", message: "skills-dir"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			output, err := runSkills(t, append([]string{"install"}, tc.args...)...)
			require.ErrorContains(t, err, tc.message)
			require.Contains(t, output, "Usage:")
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			require.Empty(t, entries)
		})
	}
}

func TestSkillsInstallReplacement(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "gh-app", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(destination), 0o755))
	require.NoError(t, os.WriteFile(destination, []byte("existing"), 0o644))
	_, err := runSkills(t, "install", "--skills-dir", dir)
	require.Error(t, err)
	content, err := os.ReadFile(destination)
	require.NoError(t, err)
	require.Equal(t, "existing", string(content))
	_, err = runSkills(t, "install", "--skills-dir", dir, "--force")
	require.NoError(t, err)
	content, err = os.ReadFile(destination)
	require.NoError(t, err)
	require.Equal(t, skills.Markdown, string(content))
}

func TestSkillsInstallFailures(t *testing.T) {
	t.Run("mkdir", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(dir, []byte("unchanged"), 0o644))
		output, err := runSkills(t, "install", "--skills-dir", dir)
		require.ErrorContains(t, err, "create skills directory")
		require.Empty(t, output)
	})
	t.Run("open", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "gh-app", "SKILL.md"), 0o755))
		output, err := runSkills(t, "install", "--skills-dir", dir, "--force")
		require.ErrorContains(t, err, "create skill")
		require.Empty(t, output)
	})
	t.Run("write", func(t *testing.T) {
		// /dev/full reliably rejects writes even when tests run as root.
		if _, err := os.Stat("/dev/full"); err != nil {
			t.Skip("requires /dev/full")
		}
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "gh-app"), 0o755))
		require.NoError(t, os.Symlink("/dev/full", filepath.Join(dir, "gh-app", "SKILL.md")))
		output, err := runSkills(t, "install", "--skills-dir", dir, "--force")
		require.ErrorContains(t, err, "write skill")
		require.Empty(t, output)
	})
}

func TestSkillsInstall(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("GHAPP_CLIENT_ID", "")
	t.Setenv("GHAPP_PRIVATE_KEY_PATH", "")
	dir := filepath.Join(t.TempDir(), "skills with spaces", "nested")
	output, err := runSkills(t, "install", "--skills-dir", dir)
	require.NoError(t, err)
	destination := filepath.Join(dir, "gh-app", "SKILL.md")
	content, err := os.ReadFile(destination)
	require.NoError(t, err)
	require.Equal(t, skills.Markdown, string(content))
	require.Contains(t, output, destination)
	stdout, err := runSkills(t)
	require.NoError(t, err)
	require.Equal(t, skills.Markdown, stdout)
}
