package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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

	var skillsDir string
	var force bool
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install SKILL.md into an explicit skills directory",
		Args:  cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if skillsDir == "" {
				return fmt.Errorf("--skills-dir must not be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			dir := filepath.Join(skillsDir, "gh-app")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("create skills directory %q: %w", dir, err)
			}
			path := filepath.Join(dir, "SKILL.md")
			flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
			if force {
				flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			}
			file, err := os.OpenFile(path, flags, 0o644)
			if err != nil {
				return fmt.Errorf("create skill %q: %w", path, err)
			}
			_, writeErr := io.WriteString(file, skills.Markdown)
			closeErr := file.Close()
			if writeErr != nil {
				return fmt.Errorf("write skill %q: %w", path, writeErr)
			}
			if closeErr != nil {
				return fmt.Errorf("close skill %q: %w", path, closeErr)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
			return err
		},
	}
	installCmd.Flags().StringVar(&skillsDir, "skills-dir", "", "Parent directory containing skills")
	installCmd.Flags().BoolVar(&force, "force", false, "Replace an existing SKILL.md")
	if err := installCmd.MarkFlagRequired("skills-dir"); err != nil {
		return nil, err
	}
	skillsCmd.AddCommand(installCmd)
	return skillsCmd, nil
}
