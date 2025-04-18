/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"
	"path"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

// strict defines whether the generate command should
//   - true: fail on missing translations
//   - false: fallback to base lang, but warn
var strict bool

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:    "generate",
	Short:  "generate published files from prepared templates and translations",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		var config illuminated.Config
		err := config.Read(path.Join(projectDir, illuminated.DefaultConfigFilename))
		if err != nil {
			slog.Error("read config", "error", err)
			os.Exit(1)
		}
		slog.Debug("config read", "config", config)
		for _, lang := range config.Targets {
			slog.Debug("generating document", "lang", lang)
			err = illuminated.Generate("downloads", lang, projectDir)
			if err != nil {
				slog.Error("generate document", "lang", lang, "error", err)
				os.Exit(1)
			}
			// TODO make real
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.PersistentFlags().BoolVarP(&strict, "strict", "s", false, "strict mode (fail on missing translations)")
	// generateCmd.MarkFlagsOneRequired("html", "pdf")
	// generateCmd.PersistentFlags().BoolVarP(&html, "html", "h", false, "generate HTML output")
	// generateCmd.PersistentFlags().BoolVarP(&pdf, "pdf", "p", false, "generate PDF output")
}
