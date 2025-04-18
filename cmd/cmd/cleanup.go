package cmd

import (
	"log/slog"
	"os"
	"path"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

var ()

var cleanupCmd = &cobra.Command{
	Use:    "cleanup",
	Short:  "deletes selected directory and all files",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("cleaning up project directory", "dir", projectDir)
		_, err := os.Stat(path.Join(projectDir, illuminated.DefaultConfigFilename))
		if err != nil {
			if os.IsNotExist(err) {
				slog.Info("no config found, nothing to clean up")
				os.Exit(0)
			}
		}
		err = os.RemoveAll(projectDir)
		if err != nil {
			slog.Error("error removing project directory", "dir", projectDir, "error", err)
			os.Exit(1)
		}
		slog.Debug("cleanup complete!")
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
}
