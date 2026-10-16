package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	t.Parallel()

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
