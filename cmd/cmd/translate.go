package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/getlantern/illuminated/cmd/translators"
	"github.com/spf13/cobra"
)

var (
	argTranslator    string
	translatorGoogle = "google"
	validTranslators = []string{translatorGoogle}
)

// translateCmd represents the translate command
var translateCmd = &cobra.Command{
	Use:    "translate",
	Short:  "Translate base language into target languages",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: load config to determine `base` and `targets`
		if slices.Contains(validTranslators, argTranslator) {
			slog.Debug("translating", "translator", argTranslator)
			translators.TranslateWithGoogle()
		} else {
			slog.Error("invalid translator specified",
				"given", argTranslator,
				"valid", validTranslators,
			)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(translateCmd)
	msg := fmt.Sprintf("%v", validTranslators)
	translateCmd.Flags().StringVarP(&argTranslator, "translator", "t", translatorGoogle, msg)
}
