package cmd

import (
	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

var (
	config     illuminated.Config
	projectDir string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize illuminated with non-default options",
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
	rootCmd.PersistentFlags().StringVarP(&projectDir, "directory", "d", "illuminated-project", "project directory for intermediate files")
	rootCmd.PersistentFlags().StringVarP(&config.Base, "base", "b", illuminated.DefaultConfig.Base, "base language for source material (ISO 639-1 codes)")
	rootCmd.PersistentFlags().StringSliceVarP(&config.Targets, "target", "t", illuminated.DefaultConfig.Targets, "target languages (ISO 639-1 codes)")
}
