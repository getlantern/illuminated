package cmd

import (
	"errors"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

var (
	configOpts  illuminated.Config
	forceReinit bool // overwrite existing config
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:    "init",
	Short:  "initialize illuminated with non-default options",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		err := configOpts.Write(projectDir, forceReinit)
		if err != nil {
			if errors.Is(err, illuminated.ErrNoClobber) {
				cmd.PrintErrf("config file already exists, use --force to overwrite")
			}
			cmd.PrintErrf("error writing config: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.PersistentFlags().BoolVarP(&forceReinit, "force", "f",
		false,
		"overwrite existing config",
	)
	initCmd.PersistentFlags().StringVarP(&configOpts.Base, "base", "b",
		illuminated.DefaultConfig.Base,
		"base language for source material (ISO 639-1 codes)",
	)
	initCmd.PersistentFlags().StringSliceVarP(&configOpts.Targets, "target", "t",
		illuminated.DefaultConfig.Targets,
		"target languages (ISO 639-1 codes)",
	)
}
