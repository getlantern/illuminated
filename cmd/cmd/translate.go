package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"

	"github.com/getlantern/illuminated"
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
		var config illuminated.Config
		err := config.Read(path.Join(projectDir, illuminated.DefaultConfigFilename))
		if err != nil {
			slog.Error("read config", "error", err)
			os.Exit(1)
		}
		slog.Debug("config read", "config", config)
		if len(config.Targets) == 0 {
			slog.Error("no target languages defined in config")
			os.Exit(1)
		}
		if config.Base == "" {
			slog.Error("no base language defined in config")
			os.Exit(1)
		}
		// conf
		// config.Read(path.Join(projectDir, illuminated.DefaultConfigFilename))

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
			defer g.Close(cmd.Context())
			langs, err := g.SuportedLanguages(cmd.Context(), "en") // TODO: get from config.Base
			if err != nil {
				slog.Error("get supported languages", "error", err)
				os.Exit(1)
			}
			for _, lang := range langs {
				slog.Debug("supported language", "language", lang)
			}
			txs, err := g.Translate(cmd.Context(), "es", []string{"Hello, world!"}) // TODO: get from config.Targets and texts to translate
			if err != nil {
				slog.Error("translate with Google", "error", err)
				os.Exit(1)
			}
			for _, tx := range txs {
				slog.Info("translated text", "text", tx)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(translateCmd)
	msg := fmt.Sprintf("%v", validTranslators)
	translateCmd.Flags().StringVarP(&argTranslator, "translator", "t", translatorGoogle, msg)
}
