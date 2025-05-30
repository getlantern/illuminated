package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/getlantern/illuminated"
	"github.com/getlantern/illuminated/cmd/translators"
	"github.com/spf13/cobra"
)

// translator is the name of the translator engine to use.
// Expected to exist in translators.ValidTranslators.
var (
	translator string
	overwrite  bool
)

var translateCmd = &cobra.Command{
	Use:    "translate",
	Short:  "Translate base language into target languages",
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: abstract this repeated code
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
		if !slices.Contains(translators.ValidTranslators, translator) {
			slog.Error("invalid translator specified",
				"given", translator,
				"valid", translators.ValidTranslators,
			)
			os.Exit(1)
		}
		slog.Debug("translating", "translator", translator)

		g, err := translators.NewTranslator(cmd.Context(), translators.GoogleTranslate)
		if err != nil {
			slog.Error("translator could not be initialized", "error", err)
			os.Exit(1)
		}

		defer g.Close(cmd.Context())
		langSupported, err := g.SupportedLanguages(cmd.Context(), config.Base)
		if err != nil {
			slog.Error("get supported languages", "error", err)
			os.Exit(1)
		}

		// read the directory once
		txPath := path.Join(projectDir, illuminated.DefaultDirNameTranslations)
		files, err := os.ReadDir(txPath)
		if err != nil {
			slog.Error("read translations directory",
				"dir", txPath,
				"error", err,
			)
			os.Exit(1)
		}

		// load the base language, for use by all targets
		baseFileTexts := make(map[string]map[string]string)
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
				continue
			}
			if strings.HasPrefix(file.Name(), config.Base) {
				f, err := os.ReadFile(path.Join(txPath, file.Name()))
				if err != nil {
					slog.Error("read base language file", "error", err)
					os.Exit(1)
				}
				var baseText map[string]string
				err = json.Unmarshal(f, &baseText)
				if err != nil {
					slog.Error("unmarshal base language file", "error", err)
					os.Exit(1)
				}
				if len(baseText) == 0 {
					slog.Error("base language file is empty or invalid", "file", file.Name())
					os.Exit(1)
				}
				name := strings.TrimSuffix(file.Name(), ".json")
				name = strings.TrimPrefix(name, config.Base+".")
				slog.Debug("base language text loaded",
					"file", file.Name(),
					"name", name,
					"textCount", len(baseText),
				)
				baseFileTexts[name] = baseText
			}
		}

		for _, target := range config.Targets {
			if !slices.Contains(langSupported, target) {
				slog.Error("target language not supported by translator",
					"target", target,
					"supported", langSupported,
				)
				os.Exit(1)
			}
			slog.Debug("translating texts", "target", target)

			// Second, we translate each target languages using the base language.
			if target == config.Base {
				slog.Debug("skipping translation for base language, assuming base already exists", "base", config.Base)
				continue
			}
			var updatedFilesQty int
			for _, file := range files {
				var targetText map[string]string
				if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
					continue
				}
				if !strings.HasPrefix(file.Name(), target+".") {
					continue
				}
				f, err := os.ReadFile(path.Join(txPath, file.Name()))
				if err != nil {
					slog.Error("read target language file", "error", err, "file", file.Name())
					os.Exit(1)
				}
				err = json.Unmarshal(f, &targetText)
				if err != nil {
					slog.Error("unmarshal target language file", "error", err, "file", file.Name())
					os.Exit(1)
				}
				if len(targetText) == 0 {
					slog.Error("target language file is empty or invalid", "file", file.Name())
					os.Exit(1)
				}
				name := strings.TrimSuffix(file.Name(), ".json")
				name = strings.TrimPrefix(name, target+".")
				slog.Debug("target language text loaded",
					"file", file.Name(),
					"name", name,
					"textCount", len(targetText),
				)
				// Perform translation
				targetStringsQty := 0
				for key, text := range baseFileTexts[name] {
					v, exists := targetText[key]
					if exists && !overwrite && v != "" {
						slog.Debug(
							"skipping existing translation to avoid clobber; use -o/--overwrite",
							"file", file.Name(),
							"existing value", v,
							"key", key,
						)
						continue
					}
					result, err := g.Translate(cmd.Context(), target, []string{text})
					if err != nil {
						slog.Error("translation failure", "error", err, "key", key, "text", text)
						os.Exit(1)
					}
					if len(result) == 0 {
						slog.Error("translation result is empty", "key", key, "text", text)
						os.Exit(1)
					}
					// NOTE: only one result is expected, forcing a line by line translation.
					// It is possible to reduce network trips by packing the tranlation map
					// into an array, but that would require handling ordering concerns,
					// while this approach guarantees completion in a single pass,
					// and skips existing fields unless -o/--overwrite is true.
					if len(result) != 1 {
						slog.Error("length error",
							"expected", 1,
							"got", len(result),
							"result", result,
						)
						os.Exit(1)
					}
					targetText[key] = result[0]
					slog.Debug("translation complete for text string",
						"key", key,
						"text", text,
						"result", targetText[key],
					)
					targetStringsQty++
				}
				outPath := path.Join(txPath, file.Name())
				content, err := json.MarshalIndent(targetText, "", "  ")
				if err != nil {
					slog.Error("marshal target language file", "error", err, "file", file.Name())
					os.Exit(1)
				}
				err = os.WriteFile(outPath, content, illuminated.DefaultFilePermissions)
				if err != nil {
					slog.Error("write target language file", "error", err, "file", file.Name())
					os.Exit(1)
				}
				updatedFilesQty++
				slog.Info("translations complete for file",
					"file", file.Name(),
					"translated", targetStringsQty,
					"total", len(baseFileTexts[name]),
				)
			}
			slog.Info("translations complete for language",
				"language", target,
				"updated", updatedFilesQty,
			)
		}
		slog.Info("all translations complete!")
	},
}

func init() {
	rootCmd.AddCommand(translateCmd)
	msg := fmt.Sprintf("%v", translators.ValidTranslators)
	translateCmd.Flags().StringVarP(&translator, "translator", "t", translators.GoogleTranslate, msg)
	translateCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "overwrite existing translations, if present")
}
