package cmd

import (
	"errors"

	"github.com/air-hand-org/gh-app/internal/dependency"
	"github.com/air-hand-org/gh-app/internal/token"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newDoctorCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check local dependencies and GitHub App configuration (offline)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			results, err := dependency.Check("gh", "git")
			for _, result := range results {
				if result.Err != nil {
					cmd.Printf("%s: not found on PATH (%v)\n", result.Name, result.Err)
				} else {
					cmd.Printf("%s: %s\n", result.Name, result.Path)
				}
			}
			checks := token.CheckConfiguration(v.GetString("client-id"), v.GetString("private-key-path"))
			failed := false
			if checks.ClientIDErr != nil {
				cmd.Printf("client-id: FAIL (%v)\n", checks.ClientIDErr)
				failed = true
			} else {
				cmd.Println("client-id: OK (configured)")
			}
			if checks.KeyFileErr != nil {
				cmd.Printf("private-key-path: FAIL (%v)\n", checks.KeyFileErr)
				cmd.Println("private-key: SKIP (file unavailable)")
				failed = true
			} else {
				cmd.Println("private-key-path: OK (readable regular file)")
				if checks.KeyErr != nil {
					cmd.Printf("private-key: FAIL (%v)\n", checks.KeyErr)
					failed = true
				} else {
					cmd.Println("private-key: OK (supported RSA private key)")
				}
			}
			if failed {
				return errors.Join(err, errors.New("local configuration checks failed"))
			}
			return err
		},
	}
}
