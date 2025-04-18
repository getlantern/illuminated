/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:    "generate",
	Short:  "generate published files from prepared templates and translations",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		// TODO don't hardcode this
		err := illuminated.Generate("downloads", "en", projectDir)
		if err != nil {
			slog.Error("generate", "error", err)
			os.Exit(1)
		}
		slog.Info("generated", "file", "downloads.html")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	// generateCmd.MarkFlagsOneRequired("html", "pdf")
	// generateCmd.PersistentFlags().BoolVarP(&html, "html", "h", false, "generate HTML output")
	// generateCmd.PersistentFlags().BoolVarP(&pdf, "pdf", "p", false, "generate PDF output")
}
