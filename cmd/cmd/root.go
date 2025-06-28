package cmd

import (
	"log/slog"
	"os"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

var verbose, quiet bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	// TODO: output markdown, too
	Short:  "make translated html or pdf from wiki or markdown",
	Long:   "fetches a local or remote source to generate HTML and/or PDF files in multiple languages.",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	RunE: func(cmd *cobra.Command, args []string) error {
		// slog.Info("illuminated CLI tool", "version", "0.1.0")
		cmd.Usage()
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVarP(
		&projectDir, "directory", "d", illuminated.DefaultDirProject,
		"project directory for output files",
	)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging (DEBUG)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet logging (ERROR)")
	rootCmd.MarkFlagsMutuallyExclusive("verbose", "quiet")
}

// Init initializes the logging level based on the flags set,
// but must be invoked in the PreRun of any subcommand, like:
//
//	PreRun: func(cmd *cobra.Command, args []string) { Init() },
func Init() {
	if quiet {
		slog.SetLogLoggerLevel(slog.LevelError)
	}
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("verbose logging enabled")
	}
}
