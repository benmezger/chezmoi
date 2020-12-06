package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

type applyCmdConfig struct {
	include   *chezmoi.IncludeSet
	recursive bool
}

func (c *Config) newApplyCmd() *cobra.Command {
	applyCmd := &cobra.Command{
		Use:     "apply [target]...",
		Short:   "Update the destination directory to match the target state",
		Long:    mustLongHelp("apply"),
		Example: example("apply"),
		RunE:    c.runApplyCmd,
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
		},
	}

	flags := applyCmd.Flags()
	flags.VarP(c.apply.include, "include", "i", "include entry types")
	flags.BoolVarP(&c.apply.recursive, "recursive", "r", c.apply.recursive, "recursive")

	return applyCmd
}

func (c *Config) runApplyCmd(cmd *cobra.Command, args []string) error {
	return c.applyArgs(c.destSystem, c.absSlashDestDir, args, c.apply.include, c.apply.recursive, c.Umask.FileMode())
}
