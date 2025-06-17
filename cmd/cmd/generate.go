package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/getlantern/illuminated"
	"github.com/getlantern/illuminated/translators"
	"github.com/spf13/cobra"
)

var (
	projectDir    string
	source        string   // source document(s): directory or GitHub wiki URL
	targetLangs   []string // target languages (ISO 639-1 codes)
	baseLang      string   // base language of source files (ISO 639-1 code)
	translator    string   // translator to use, e.g. "google", "mock"
	overridesPath string   // path to yaml file defining overrides
	join          bool     // join HTML files into single document or split into individual files?
	html          bool     // generate HTML output
	pdf           bool     // generate PDF output
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate documents from source files",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:
	//
	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	RunE: func(cmd *cobra.Command, args []string) error {
		// stage files from remote or outside dir to projectDir
		err := illuminated.Stage(source, projectDir)
		if err != nil {
			slog.Error("unable to stage selected source", "error", err)
			os.Exit(1)
		}
		slog.Debug("source files staged", "source", source, "projectDir", projectDir)

		// process the markdown into html
		files, err := os.ReadDir(path.Join(projectDir, illuminated.DefaultDirNameStaging))
		if err != nil {
			return fmt.Errorf("read staging directory: %w", err)
		}

		// ensure output directory exists
		err = os.MkdirAll(
			path.Join(projectDir, illuminated.DefaultDirNameOutput),
			illuminated.DefaultFilePermissions,
		)
		if err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}

		var g translators.Translator
		if len(targetLangs) > 0 {
			g, err = translators.NewTranslator(cmd.Context(), translator)
			if err != nil {
				return fmt.Errorf("create %q translator client: %w", translator, err)
			}
		}

		// generate HTML from markdown
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				slog.Debug("skipping dir (expects only markdown files)", "name", file.Name())
				continue
			}
			sourcePath := filepath.Join(projectDir, illuminated.DefaultDirNameStaging, file.Name())
			slog.Debug("reading markdown file", "path", sourcePath)
			outName := strings.TrimSuffix(file.Name(), ".md")
			outName = fmt.Sprintf("%s.%s.%s", baseLang, outName, "html")
			outPath := path.Join(projectDir, illuminated.DefaultDirNameOutput, outName)

			err := illuminated.MarkdownToHTML(sourcePath, outPath)
			if err != nil {
				return fmt.Errorf("reading markdown file %q: %w", sourcePath, err)
			}
			slog.Debug("created HTML in base lang",
				"source", sourcePath,
				"lang", baseLang,
				"out", outPath,
			)

			// also generate HTML for each target language
			for _, lang := range targetLangs {
				if lang == baseLang {
					continue
				}
				outName = strings.TrimSuffix(file.Name(), ".md")
				outName = strings.TrimPrefix(outName, baseLang+".")
				txOutName := fmt.Sprintf("%s.%s.%s", lang, outName, "html")
				txOutPath := path.Join(projectDir, illuminated.DefaultDirNameOutput, txOutName)

				baseLangFileData, err := os.ReadFile(outPath)
				if err != nil {
					return fmt.Errorf("read base language file %q: %w", outPath, err)
				}
				tx, err := g.Translate(cmd.Context(), lang, []string{string(baseLangFileData)})
				if err != nil || len(tx) == 0 {
					return fmt.Errorf("translate file %q to language %q: %w", outPath, lang, err)
				}

				// apply any overrides
				if overridesPath == "" {
					overridesPath = path.Join(illuminated.DefaultFileNameOverrides)
				}
				overrides, err := illuminated.ReadOverrideFile(illuminated.DefaultFileNameOverrides)
				if err != nil {
					if !os.IsNotExist(err) {
						slog.Debug("no override file found",
							"expected", overridesPath,
						)
					} else {
						return fmt.Errorf("read override file %q: %w", overridesPath, err)
					}
				}
				for _, override := range overrides {
					if override.Language != lang {
						continue
					}
					if override.Original == "" || override.Replacement == "" {
						slog.Warn("skipping override with empty original or replacement",
							"override", override,
						)
						continue
					}
					for i, t := range tx {
						tx[i] = strings.ReplaceAll(t, override.Original, override.Replacement)
						slog.Debug("applied override",
							"title", override.Title,
							"original", override.Original,
							"replacement", override.Replacement,
							"lang", override.Language,
							"file", txOutName,
						)
					}
				}

				err = os.WriteFile(txOutPath, []byte(tx[0]), illuminated.DefaultFilePermissions)
				if err != nil {
					return fmt.Errorf("write translated file %q: %w", txOutPath, err)
				}
				slog.Debug("created HTML in target language",
					"source", outPath,
					"target", txOutPath,
					"lang", lang,
				)
			}
		}

		// join all files for a language into one HTML
		if join {
			// join all HTML files into one
			for _, lang := range targetLangs {
				joinedFile, err := illuminated.JoinHTML(lang, projectDir, projectDir)
				if err != nil {
					return fmt.Errorf("join HTML files for language %q: %w", lang, err)
				}
				slog.Debug("joined HTML files", "file", joinedFile)
			}
		}

		// generate PDF files from HTML
		if pdf {
			slog.Debug("generating pdf")
			files, err := os.ReadDir(path.Join(projectDir, illuminated.DefaultDirNameOutput))
			if err != nil {
				return fmt.Errorf("read output directory: %w", err)
			}
			for _, file := range files {
				if file.IsDir() || !strings.HasSuffix(file.Name(), ".html") {
					slog.Debug("skipping dir (expects only HTML files)", "name", file.Name())
					continue
				}
				sourcePath := filepath.Join(projectDir, illuminated.DefaultDirNameOutput, file.Name())
				// generate PDF files from HTML
				parts := strings.Split(file.Name(), ".")
				if len(parts) < 3 {
					return fmt.Errorf("invalid file name %q, expected format: <lang>.<name>.html", file.Name())
				}
				lang := parts[0]
				var name string
				if join {
					name = projectDir
				} else {
					name = file.Name()
				}
				outName := fmt.Sprintf("%s.%s.pdf", lang, name)
				outPath := path.Join(projectDir, illuminated.DefaultDirNameOutput, outName)
				if _, err := os.Stat(outPath); !os.IsNotExist(err) {
					if !force {
						// TODO: this may be throwing false positives
						slog.Info("skipping file to avoid clobber, set -f/--force to overwrite", "file", outPath)
						continue
					}
				}
				resources := path.Join(projectDir, illuminated.DefaultDirNameStaging)
				err := illuminated.WritePDF(sourcePath, outPath, resources)
				if err != nil {
					return fmt.Errorf("generate PDF for lang %q: %w", baseLang, err)
				}
				// remove HTML if only used as intermediate file for PDF generation
				if !html {
					slog.Debug("removing HTML file after PDF generation", "file", sourcePath)
					err = os.Remove(sourcePath)
					if err != nil {
						return fmt.Errorf("remove HTML file after PDF generation: %w", err)
					}
				}
			}
		}
		slog.Info("document generation complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// source
	generateCmd.PersistentFlags().StringVarP(
		&source, "source", "s", "",
		"source document(s) location, can be: directory, or GitHub wiki URL",
	)
	generateCmd.MarkPersistentFlagRequired("source")

	// translation
	generateCmd.PersistentFlags().StringVarP(
		&baseLang, "base", "b", "en",
		"language (ISO 639-1 code) of source files",
	)
	generateCmd.PersistentFlags().StringSliceVarP(
		&targetLangs, "languages", "l", []string{},
		"target languages to translated from source (ISO 639-1 codes)",
	)
	generateCmd.PersistentFlags().StringVarP(&translator, "translator", "t", "", "translator service to use")
	// NOTE: when base=languages, this is noop, but still required
	generateCmd.MarkFlagsRequiredTogether("languages", "translator")
	generateCmd.PersistentFlags().StringVarP(
		&overridesPath, "overrides", "o",
		path.Join(illuminated.DefaultFileNameOverrides),
		"path to yaml file defining overrides, see readme for example",
	)

	// output
	generateCmd.PersistentFlags().BoolVarP(&join, "join", "j", false, "join all documents into one")
	generateCmd.PersistentFlags().BoolVarP(&html, "html", "H", false, "generate HTML output")
	generateCmd.PersistentFlags().BoolVarP(&pdf, "pdf", "P", false, "generate PDF output")
	generateCmd.MarkFlagsOneRequired("html", "pdf")
	generateCmd.PersistentFlags().BoolVarP(&force, "force", "f",
		false,
		"overwrite existing files",
	)
}
