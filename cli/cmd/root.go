package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	source  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Short: "generate translated PDF from markdown source file(s)",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging (DEBUG)")
}

func Init() {
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("verbose logging enabled")
	}
}
