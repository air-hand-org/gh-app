package cmd

import (
	"io"

	"github.com/air-hand-org/gh-app/internal/skills"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newSkillsCmd(_ *viper.Viper) (*cobra.Command, error) {
	var skillsCmd = &cobra.Command{
		Use:   "skills",
		Short: "output SKILL.md to stdout",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := io.WriteString(cmd.OutOrStdout(), skills.Markdown)
			return err
		},
	}

	return skillsCmd, nil
}
