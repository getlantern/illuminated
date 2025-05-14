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

// updateCmd fetches new source material and updates the project directory with
// new translation .json files and HTML templates.
var updateCmd = &cobra.Command{
	Use:    "update",
	Short:  "stage source files, parse them, and create templates and translation files",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		// stage files from remote or outside dir to projectDir
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
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				slog.Debug("skipping dir (expects only files)", "name", file.Name())
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
	rootCmd.AddCommand(updateCmd)
	updateCmd.PersistentFlags().StringVarP(&source, "source", "s", "", "source document(s) location, can be: directory, or GitHub wiki URL")
	updateCmd.MarkPersistentFlagRequired("source")
}
