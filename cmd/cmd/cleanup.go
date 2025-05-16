package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/getlantern/illuminated"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var force bool

var cleanupCmd = &cobra.Command{
	Use:    "cleanup",
	Short:  "deletes selected directory and all files",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		if !force {
			p := promptui.Prompt{
				Label:     fmt.Sprintf("Are you sure you want to delete the project directory %q and everything in it? (yes/no)", projectDir),
				IsConfirm: true,
			}
			_, err := p.Run()
			if err != nil {
				slog.Error("confirmation or --force required", "error", err)
				os.Exit(1)
			}
		}

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
	cleanupCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "force cleanup without confirmation")
}
