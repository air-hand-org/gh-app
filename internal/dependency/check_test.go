package dependency

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gh")
	require.NoError(t, os.WriteFile(path, []byte("fixture\n"), 0o755))
	t.Setenv("PATH", dir)
	results, err := Check("gh", "git", "missing")
	require.Error(t, err)
	require.Len(t, results, 3)
	require.Equal(t, "gh", results[0].Name)
	require.Equal(t, path, results[0].Path)
	require.NoError(t, results[0].Err)
	require.Equal(t, "git", results[1].Name)
	require.Error(t, results[1].Err)
	require.Error(t, results[2].Err)
	require.ErrorContains(t, err, "git")
	require.ErrorContains(t, err, "missing")
}

func TestCheckSuccess(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"gh", "git"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("fixture\n"), 0o755))
	}
	t.Setenv("PATH", dir)
	results, err := Check("gh", "git")
	require.NoError(t, err)
	require.Len(t, results, 2)
	for _, result := range results {
		require.Equal(t, filepath.Join(dir, result.Name), result.Path)
		require.NoError(t, result.Err)
	}
}
