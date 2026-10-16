package cmd

import (
	"os/exec"

	"github.com/cockroachdb/errors"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check that gh and git are available on PATH",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var checkErr error
			for _, name := range []string{"gh", "git"} {
				path, err := exec.LookPath(name)
				if err != nil {
					cmd.Printf("%s: not found on PATH (%v)\n", name, err)
					checkErr = errors.Join(checkErr, err)
					continue
				}
				cmd.Printf("%s: %s\n", name, path)
			}
			return checkErr
		},
	}
}
