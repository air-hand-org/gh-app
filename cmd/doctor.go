package cmd

import (
	"github.com/air-hand-org/gh-app/internal/dependency"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check that gh and git are available on PATH",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			results, err := dependency.Check("gh", "git")
			for _, result := range results {
				if result.Err != nil {
					cmd.Printf("%s: not found on PATH (%v)\n", result.Name, result.Err)
				} else {
					cmd.Printf("%s: %s\n", result.Name, result.Path)
				}
			}
			return err
		},
	}
}
