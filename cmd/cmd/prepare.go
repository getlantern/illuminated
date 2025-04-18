package cmd

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

// prepareCmd represents the prepare command
var prepareCmd = &cobra.Command{
	Use:    "prepare",
	Short:  "stage source files, parse them, and create templates and translation files",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		// stage
		err := illuminated.Stage(source, projectDir)
		if err != nil {
			slog.Error("unable to stage selected source", "error", err)
			os.Exit(1)
		}
		slog.Debug("source files staged")

		// process
		files, err := os.ReadDir(path.Join(projectDir, illuminated.DefaultDirNameStaging))
		if err != nil {
			slog.Error("read staging directory", "error", err)
			os.Exit(1)
		}
		for _, file := range files {
			if file.IsDir() {
				slog.Warn("skipping dir (expects only files)", "name", file.Name())
				continue
			}
			filePath := filepath.Join(projectDir, illuminated.DefaultDirNameStaging, file.Name())
			slog.Debug("processing file", "file", filePath)
			err := illuminated.Process(filePath, projectDir)
			if err != nil {
				slog.Error("process file", "file", filePath, "error", err)
				os.Exit(1)
			}
			slog.Debug("processed document", "doc", filePath)
		}
	},
}

func init() {
	rootCmd.AddCommand(prepareCmd)
	prepareCmd.PersistentFlags().StringVarP(&source, "source", "s", "", "source document(s) location, can be: file, directory, or GitHub wiki URL")
	prepareCmd.MarkPersistentFlagRequired("source")
}
