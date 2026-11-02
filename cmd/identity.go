package cmd

import (
	"context"
	"encoding/json"

	"github.com/air-hand-org/gh-app/internal/token"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type IdentityResolver interface {
	Resolve(ctx context.Context, clientID, privateKeyPath string) (token.Identity, error)
}

func newIdentityCmd(v *viper.Viper, resolver IdentityResolver) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use: "identity", Short: "Show the authenticated GitHub App bot's commit identity", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			identity, err := resolver.Resolve(cmd.Context(), v.GetString("client-id"), v.GetString("private-key-path"))
			if err != nil {
				return err
			}
			if asJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(identity)
			}
			_, err = cmd.OutOrStdout().Write([]byte(identity.Name + " <" + identity.Email + ">\n"))
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print identity as JSON")
	return cmd
}
