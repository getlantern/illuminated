package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose, quiet bool
	projectDir     string
	source         string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Short: "prepare templates and translations files from markdown, then generate translated PDF files",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVarP(&projectDir, "directory", "d", "illuminated-project", "project directory for intermediate files")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging (DEBUG)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet logging (ERROR)")
	rootCmd.MarkFlagsMutuallyExclusive("verbose", "quiet")
}

func Init() {
	if quiet {
		slog.SetLogLoggerLevel(slog.LevelError)
	}
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("verbose logging enabled")
	}
}
