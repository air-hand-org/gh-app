// Command gh-app provides the CLI entry point for GitHub App authentication.
package main

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	command := &cobra.Command{
		Use:     "gh-app",
		Short:   "Run GitHub CLI with GitHub App authentication (implementation pending)",
		Version: version,
		Args:    cobra.NoArgs,
	}
	command.CompletionOptions.DisableDefaultCmd = true
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
