package cmd

import (
	"log/slog"
	"os"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

var (
	staging, output, templates, translations bool
)

var cleanupCmd = &cobra.Command{
	Use:    "cleanup",
	Short:  "deletes selected directory and all files",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		if staging {
			err := os.RemoveAll(illuminated.DefaultDirNameStaging)
			if err != nil {
				slog.Error("unable to cleanup", "dir", illuminated.DefaultDirNameStaging, "error", err)
				os.Exit(1)
			}
		}
		if output {
			err := os.RemoveAll(illuminated.DefaultDirNameOutput)
			if err != nil {
				slog.Error("unable to cleanup", "dir", illuminated.DefaultDirNameOutput, "error", err)
				os.Exit(1)
			}
		}
		if templates {
			err := os.RemoveAll(illuminated.DefaultDirNameTemplates)
			if err != nil {
				slog.Error("unable to cleanup", "dir", illuminated.DefaultDirNameTemplates, "error", err)
				os.Exit(1)
			}
		}
		if translations {
			err := os.RemoveAll(illuminated.DefaultDirNameTranslations)
			if err != nil {
				slog.Error("unable to cleanup", "dir", illuminated.DefaultDirNameTranslations, "error", err)
				os.Exit(1)
			}
		}
		slog.Debug("cleanup complete!")
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.PersistentFlags().BoolVarP(&staging, "staging", "g", true, "delete staging directory")
	cleanupCmd.PersistentFlags().BoolVarP(&templates, "templates", "t", false, "delete templates directory")
	cleanupCmd.PersistentFlags().BoolVarP(&translations, "translations", "l", false, "delete translations directory")
	cleanupCmd.PersistentFlags().BoolVarP(&output, "output", "o", false, "delete output directory")
}
