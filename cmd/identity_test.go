package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/air-hand-org/gh-app/internal/token"
	"github.com/stretchr/testify/require"
)

type identityResolverFunc func(context.Context, string, string) (token.Identity, error)

func (f identityResolverFunc) Resolve(ctx context.Context, id, path string) (token.Identity, error) {
	return f(ctx, id, path)
}

func TestIdentityCommand(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("GHAPP_CLIENT_ID", "environment-client")
	t.Setenv("GHAPP_PRIVATE_KEY_PATH", "environment-path")
	cmd, err := newRootCmd(rootDeps{identityResolver: identityResolverFunc(func(_ context.Context, id, path string) (token.Identity, error) {
		require.Equal(t, "flag-client", id)
		require.Equal(t, "environment-path", path)
		return token.Identity{ID: 456, Login: "app[bot]", Name: "app[bot]", Email: "456+app[bot]@users.noreply.github.com"}, nil
	})})
	require.NoError(t, err)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"identity", "--json", "--client-id", "flag-client"})
	require.NoError(t, cmd.ExecuteContext(t.Context()))
	require.JSONEq(t, `{"id":456,"login":"app[bot]","name":"app[bot]","email":"456+app[bot]@users.noreply.github.com"}`, output.String())
}
