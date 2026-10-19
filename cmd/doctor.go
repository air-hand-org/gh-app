package cmd

import (
	"errors"
	"os"

	"github.com/air-hand-org/gh-app/internal/dependency"
	"github.com/jferrl/go-githubauth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// readDoctorKey returns only fixed diagnostics: filesystem errors may contain sensitive paths.
func readDoctorKey(path string) ([]byte, string) {
	if path == "" {
		return nil, "not configured"
	}
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, "file does not exist"
	}
	if err != nil {
		return nil, "file is not accessible"
	}
	if !info.Mode().IsRegular() {
		return nil, "not a regular file"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "file is not readable"
	}
	return data, ""
}

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
			failed := false
			if v.GetString("client-id") == "" {
				cmd.Println("client-id: FAIL (not configured)")
				failed = true
			} else {
				cmd.Println("client-id: OK (configured)")
			}
			data, reason := readDoctorKey(v.GetString("private-key-path"))
			if reason != "" {
				cmd.Printf("private-key-path: FAIL (%s)\n", reason)
				cmd.Println("private-key: SKIP (file unavailable)")
				failed = true
			} else {
				cmd.Println("private-key-path: OK (readable regular file)")
				// Use the authentication library's parser without calling Token:
				// construction is offline and does not generate a JWT. A fixed
				// issuer keeps key validation independent of missing Client ID.
				if _, keyErr := githubauth.NewApplicationTokenSource("doctor", data); keyErr != nil {
					cmd.Println("private-key: FAIL (invalid or unsupported private key)")
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
