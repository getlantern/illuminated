/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

// prepareCmd represents the prepare command
var prepareCmd = &cobra.Command{
	Use:    "prepare",
	Short:  "stage source files, generate templates and translation files",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		// stage
		err := illuminated.Stage(source)
		if err != nil {
			slog.Error("unable to stage selected source", "error", err)
			os.Exit(1)
		}
		slog.Info("source files staged")

		// process
		files, err := os.ReadDir(illuminated.DirStaging)
		if err != nil {
			slog.Error("read staging directory", "error", err)
			os.Exit(1)
		}
		for _, file := range files {
			if file.IsDir() {
				slog.Warn("skipping dir (expects only files)", "name", file.Name())
				continue
			}
			filePath := filepath.Join(illuminated.DirStaging, file.Name())
			slog.Debug("processing file", "file", filePath)
			doc, strs, err := illuminated.Process(filePath)
			if err != nil {
				slog.Error("process file", "file", filePath, "error", err)
				os.Exit(1)
			}
			slog.Info("processed", "doc", filePath, "strings", len(strs), "docaddr", doc)
			// fmt.Println(doc)
			// fmt.Println(strs)

			// TODO

		}
		// - html.tmpl
		// - .json translation file

		// illuminated.Do()

	},
}

func init() {
	rootCmd.AddCommand(prepareCmd)
	prepareCmd.PersistentFlags().StringVarP(&source, "source", "s", "", "source document(s) location, can be: file, directory, or GitHub wiki URL")
	prepareCmd.MarkPersistentFlagRequired("source")
}
