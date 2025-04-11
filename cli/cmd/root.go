package cmd

import (
	"log/slog"
	"os"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

var (
	verbose, cleanup bool
	source           string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cmd",
	Short: "Create multiple",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}
		if cleanup {
			defer func() {
				err := os.RemoveAll(illuminated.StagingDir)
				if err != nil {
					slog.Error("unable to cleanup staging dir")
				} else {
					slog.Debug("staging dir cleaned up")
				}
			}()
		}
		slog.Debug("staging...")
		err := illuminated.Stage(source)
		if err != nil {
			slog.Error("unable to stage selected source", "error", err)
			os.Exit(1)
		}

		slog.Debug("processing...")

		illuminated.Do()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging (DEBUG)")

	rootCmd.PersistentFlags().StringVarP(&source, "source", "s", "", "source document(s) location, can be: file, directory, or GitHub wiki URL")
	rootCmd.MarkPersistentFlagRequired("source")

	rootCmd.PersistentFlags().BoolVarP(&cleanup, "cleanup", "c", false, "cleanup staging directory after completion")
}
