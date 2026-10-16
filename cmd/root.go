package cmd

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/air-hand-org/gh-app/internal/dependency"
	"github.com/air-hand-org/gh-app/internal/gh"
	"github.com/air-hand-org/gh-app/internal/token"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "dev"

//go:generate moq -out gen_owner_resolver_moq_test.go . OwnerResolver
type OwnerResolver interface {
	Owner(ctx context.Context, explicitOwner string) (string, error)
}

//go:generate moq -out gen_token_generator_moq_test.go . TokenGenerator
type TokenGenerator interface {
	New(ctx context.Context, owner string, clientID string, privateKeyPath string) (string, error)
}

//go:generate moq -out gen_ghcli_executor_moq_test.go . GHCliExecutor
type GHCliExecutor interface {
	Exec(ctx context.Context, token string, args []string) error
}

type rootDeps struct {
	ownerResolver  OwnerResolver
	tokenGenerator TokenGenerator
	ghcliExecutor  GHCliExecutor
}

func newRootCmd(deps rootDeps) (*cobra.Command, error) {
	v := viper.New()

	var rootCmd = &cobra.Command{
		Use:     "gh-app [flags] -- [gh args...]",
		Short:   "Run GitHub CLI with GitHub App authentication (implementation pending)",
		Version: version,
		Args: func(cmd *cobra.Command, _ []string) error {
			n := cmd.ArgsLenAtDash()
			if n < 0 {
				return errors.New("Specify that args for gh after '--' ")
			}
			if n > 0 {
				return errors.New("Do not pass position args before '--' ")
			}
			return nil
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			names := []string{"gh"}
			if v.GetString("owner") == "" {
				names = append(names, "git")
			}
			_, err := dependency.Check(names...)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, err := deps.ownerResolver.Owner(cmd.Context(), v.GetString("owner"))
			if err != nil {
				return err
			}

			token, err := deps.tokenGenerator.New(cmd.Context(), owner, v.GetString("client-id"), v.GetString("private-key-path"))
			if err != nil {
				return err
			}

			return deps.ghcliExecutor.Exec(cmd.Context(), token, args)
		},
	}

	if err := initializeFlags(rootCmd, v); err != nil {
		return nil, err
	}

	skillsCmd, err := newSkillsCmd(v)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(skillsCmd, newDoctorCmd())

	return rootCmd, nil
}

func initializeFlags(cmd *cobra.Command, v *viper.Viper) error {
	// Global options
	cmd.PersistentFlags().String("owner", "", "Specify target owner (org|user) account")

	// Command specific options
	cmd.Flags().String("client-id", "", "GitHub App Client ID")
	cmd.Flags().String("private-key-path", "", "GitHub App Private key file path")

	v.SetEnvPrefix("GHAPP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if err := v.BindPFlags(cmd.PersistentFlags()); err != nil {
		return err
	}
	return v.BindPFlags(cmd.Flags())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd, err := newRootCmd(rootDeps{
		ownerResolver:  gh.NewOwnerResolver(),
		tokenGenerator: token.NewInstallationTokenGenerator(),
		ghcliExecutor:  gh.NewGHCliExecutor(),
	})
	if err != nil {
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
