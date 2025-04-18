package cmd

import (
	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

var (
	config illuminated.Config
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:    "init",
	Short:  "initialize illuminated with non-default options",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		err := illuminated.DefaultConfig.Write(projectDir)
		if err != nil {
			cmd.PrintErrf("error writing config: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.PersistentFlags().StringVarP(&config.Base, "base", "b", illuminated.DefaultConfig.Base, "base language for source material (ISO 639-1 codes)")
	initCmd.PersistentFlags().StringSliceVarP(&config.Targets, "target", "t", illuminated.DefaultConfig.Targets, "target languages (ISO 639-1 codes)")
}
