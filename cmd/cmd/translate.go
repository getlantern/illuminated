package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/getlantern/illuminated/cmd/translators"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
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
		if !slices.Contains(validTranslators, argTranslator) {
			slog.Error("invalid translator specified",
				"given", argTranslator,
				"valid", validTranslators,
			)
			os.Exit(1)
		}
		slog.Debug("translating", "translator", argTranslator)
		switch argTranslator {
		case translatorGoogle:
			g, err := translators.NewGoogleTranslator(cmd.Context())
			if err != nil {
				slog.Error("create Google translator", "error", err)
				os.Exit(1)
			}
			defer g.Client.Close()

			obj, err := g.Client.SupportedLanguages(cmd.Context(), language.English)
			if err != nil {
				slog.Error("get supported languages", "error", err)
				os.Exit(1)
			}
			for _, lang := range obj {
				slog.Debug("supported language", "language", lang)
			}

			err = g.TranslateWithGoogle(cmd.Context())
			if err != nil {
				slog.Error("translate with Google", "error", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(translateCmd)
	msg := fmt.Sprintf("%v", validTranslators)
	translateCmd.Flags().StringVarP(&argTranslator, "translator", "t", translatorGoogle, msg)
}
