package cmd

import (
	"os/exec"
	"strings"

	"github.com/air-hand-org/gh-app/internal/dependency"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newGitConfigCheckCmd(v *viper.Viper, resolver IdentityResolver) *cobra.Command {
	return &cobra.Command{
		Use: "check", Short: "Check effective Git user.name against the authenticated App bot (read-only)", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			if _, err := dependency.Check("git"); err != nil {
				return err
			}
			git := exec.CommandContext(cmd.Context(), "git", "config", "--get", "user.name")
			git.Stderr = cmd.ErrOrStderr()
			output, err := git.Output()
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
					return errors.New("Git user.name is unset; run gh-app git-config to configure the App bot")
				}
				return errors.Wrap(err, "read effective Git user.name")
			}
			identity, err := resolver.Resolve(cmd.Context(), v.GetString("client-id"), v.GetString("private-key-path"))
			if err != nil {
				return err
			}
			name := strings.TrimSuffix(string(output), "\n")
			if name != identity.Name {
				return errors.Newf("Git user.name %q does not match authenticated App bot %q; run gh-app git-config to update repository-local identity", name, identity.Name)
			}
			cmd.Printf("Git user.name matches authenticated App bot: %s\n", name)
			return nil
		},
	}
}
