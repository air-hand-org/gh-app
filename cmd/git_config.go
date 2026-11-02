package cmd

import (
	"os/exec"

	"github.com/air-hand-org/gh-app/internal/dependency"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newGitConfigCmd(v *viper.Viper, resolver IdentityResolver) *cobra.Command {
	var local, global bool
	cmd := &cobra.Command{
		Use: "git-config", Short: "Set Git user.name and user.email to the authenticated App bot", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			if _, err := dependency.Check("git"); err != nil {
				return err
			}
			scope := "--local"
			if global {
				scope = "--global"
			}
			// Fail before authentication when local configuration has no repository.
			if !global {
				if err := exec.CommandContext(cmd.Context(), "git", "rev-parse", "--git-dir").Run(); err != nil {
					return errors.Wrap(err, "local git-config requires a Git repository")
				}
			}
			identity, err := resolver.Resolve(cmd.Context(), v.GetString("client-id"), v.GetString("private-key-path"))
			if err != nil {
				return err
			}
			for _, setting := range []struct{ key, value string }{{"user.name", identity.Name}, {"user.email", identity.Email}} {
				git := exec.CommandContext(cmd.Context(), "git", "config", scope, "--replace-all", setting.key, setting.value)
				git.Stderr = cmd.ErrOrStderr()
				if err := git.Run(); err != nil {
					return errors.Wrapf(err, "set Git %s", setting.key)
				}
			}
			cmd.Printf("Set Git identity (%s): %s <%s>\n", scope[2:], identity.Name, identity.Email)
			return nil
		},
	}
	cmd.Flags().BoolVar(&local, "local", false, "Write repository-local Git config (default)")
	cmd.Flags().BoolVar(&global, "global", false, "Write global Git config")
	cmd.MarkFlagsMutuallyExclusive("local", "global")
	return cmd
}
