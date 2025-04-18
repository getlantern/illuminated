/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

var (
	// strict defines whether the generate command should
	//   - true: fail on missing translations
	//   - false: fallback to base lang, but warn
	strict bool
	// join HTML files into single document or split into individual files?
	join      bool
	html, pdf bool
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:    "generate",
	Short:  "generate published files from prepared templates and translations",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	RunE: func(cmd *cobra.Command, args []string) error {
		// load and validate config
		var config illuminated.Config
		err := config.Read(path.Join(projectDir, illuminated.DefaultConfigFilename))
		if err != nil {
			slog.Error("read config", "error", err)
			os.Exit(1)
		}
		slog.Debug("config read", "config", config)
		if len(config.Targets) == 0 {
			slog.Error("no target languages defined in config")
			os.Exit(1)
		}
		if config.Base == "" {
			slog.Error("no base language defined in config")
			os.Exit(1)
		}

		err = illuminated.GenerateHTMLs(config.Base, config.Targets, projectDir, strict)
		if err != nil {
			return fmt.Errorf("generate HTMLs: %v", err)
		}

		// FEATURE: take name on input instead of defaulting to projectDir
		name := projectDir

		var joinedFilePath string
		if join {
			for _, lang := range config.Targets {
				joinedFilePath, err = illuminated.JoinHTML(lang, projectDir, name)
				if err != nil {
					return fmt.Errorf("join HTMLs: %v", err)
				}
			}
		}
		if html {
			return nil
		}

		err = illuminated.WritePDF(
			path.Join(joinedFilePath),
			path.Join(projectDir, illuminated.DefaultDirNameOutput, name+".pdf"),
		)
		if err != nil {
			return fmt.Errorf("write PDF: %v", err)
		}
		// err = os.Remove(joinedFilePath)
		// if err != nil {
		// 	return fmt.Errorf("remove joined HTML after PDF creation: %v", err)
		// }
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.PersistentFlags().BoolVarP(&strict, "strict", "s", false, "strict mode (fail on missing translations)")
	generateCmd.PersistentFlags().BoolVarP(&join, "join", "j", false, "join all documents into one")
	generateCmd.PersistentFlags().BoolVarP(&html, "html", "m", false, "generate HTML output")
	generateCmd.PersistentFlags().BoolVarP(&pdf, "pdf", "p", false, "generate PDF output")
	generateCmd.MarkFlagsOneRequired("html", "pdf")
}
