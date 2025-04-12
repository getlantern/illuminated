/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:    "generate",
	Short:  "generate published files from prepared templates and translations",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		err := illuminated.Generate("downloads", "en")
		if err != nil {
			slog.Error("generate", "error", err)
			os.Exit(1)
		}
		slog.Info("generated", "file", "downloads.html")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
