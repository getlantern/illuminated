/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

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

		// load config
		// if strict, verify that every language in config.Targets exists in translations

		// look for and parse Home.md as TOC, or just use alphabetical

		// generate all docs consolidated into one doc

		for _, lang := range config.Targets {
			slog.Debug("generating documents for language", "lang", lang)

			err = filepath.Walk(path.Join(projectDir, illuminated.DefaultDirNameTranslations), func(p string, info os.FileInfo, err error) error {
				slog.Info("walking project directory", "path", p)
				if err != nil {
					slog.Error("walk project directory", "error", err)
					return err
				}
				if info.IsDir() {
					return nil
				}
				name := path.Base(p)
				name = strings.TrimPrefix(name, lang+".")
				name = strings.TrimSuffix(name, ".json")
				slog.Info("generating document", "file", p)
				err = illuminated.Generate(name, lang, projectDir)
				return err
			})
			if err != nil {
				slog.Error("generate documents", "error", err)
				os.Exit(1)
			}
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
