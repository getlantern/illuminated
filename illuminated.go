// package converts markdown into corresponding templates and translations,
// optionally also generating rendered, translated final outputs from completed translations.
//
// Translation is outside the scope of this package.
package illuminated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/russross/blackfriday/v2"
	"golang.org/x/net/html"
)

// Process an input markdown file into parts:
//   - <DirTranslations>/<lang>.json       (translation strings for base and target languages)
//   - <DirTranslations>/<file>.html.tmpl  (go template)
func Process(input string, projectDir string) error {
	var counter int
	baseLangStrings := make(map[string]string)
	doc, err := parseHTML(input)
	if err != nil {
		return fmt.Errorf("parse input: %w", err)
	}
	extractInnerHTML(doc, baseLangStrings, &counter)
	for k, v := range baseLangStrings {
		slog.Debug("extracted", "key", k, "value", v, "file", input)
	}
	docBaseName := strings.TrimSuffix(path.Base(input), path.Ext(input))

	// load and validate config
	var config Config
	err = config.Read(path.Join(projectDir, DefaultConfigFilename))
	if err != nil {
		slog.Error("read config", "error", err)
		os.Exit(1)
	}
	slog.Debug("config read", "config", config)

	// make json (base language)
	dirTranslations := path.Join(projectDir, DefaultDirNameTranslations)
	err = os.MkdirAll(dirTranslations, DefaultFilePermissions)
	if err != nil {
		return fmt.Errorf("create directory %q: %w", dirTranslations, err)
	}
	jsonOut := path.Join(dirTranslations, fmt.Sprintf("%s.%s.json", config.Base, docBaseName))
	err = writeJSON(jsonOut, baseLangStrings)
	if err != nil {
		return fmt.Errorf("write %v: %w", jsonOut, err)
	}
	slog.Debug("translation strings written for base language", "file", jsonOut)

	// empty base language strings for other languages
	emptyStrings := make(map[string]string)
	for k := range baseLangStrings {
		emptyStrings[k] = ""
	}

	// make json (target languages)
	for _, lang := range config.Targets {
		jsonOut := path.Join(dirTranslations, fmt.Sprintf("%s.%s.json", lang, docBaseName))
		var data map[string]string
		if lang == config.Base {
			data = baseLangStrings
		} else {
			data = emptyStrings
		}
		err = writeJSON(jsonOut, data)
		if err != nil {
			return fmt.Errorf("write %v: %w", jsonOut, err)
		}
		slog.Debug("translation strings written for target language",
			"file", jsonOut,
			"lang", lang,
		)
	}

	// template
	dirTemplates := path.Join(projectDir, DefaultDirNameTemplates)
	err = os.MkdirAll(dirTemplates, DefaultFilePermissions)
	if err != nil {
		return fmt.Errorf("create directory %v: %w", dirTemplates, err)
	}
	tmplOut := path.Join(dirTemplates, fmt.Sprintf("%s.html.tmpl", docBaseName))
	err = writeHTML(tmplOut, doc)
	if err != nil {
		return fmt.Errorf("write HTML template: %w", err)
	}
	slog.Debug("HTML template written", "file", tmplOut)

	return nil
}

// extractInnerHTML extracts innerHTML strings into a map and
// replaces innerHTML with placeholders for internationalization.
func extractInnerHTML(n *html.Node, text map[string]string, counter *int) {
	if n.Type == html.TextNode {
		if len(strings.TrimSpace(n.Data)) > 0 {
			*counter++                               // increment field number...
			key := fmt.Sprintf("key_%02d", *counter) // to use as key for translation values,
			text[key] = n.Data                       // capture into map for translation file, and
			n.Data = fmt.Sprintf("{{ .%s }}", key)   // replace innerHTML with template placeholder
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractInnerHTML(c, text, counter)
	}
}

// parseHTML converts markdown file into an HTML object.
func parseHTML(inputPath string) (*html.Node, error) {
	f, err := os.ReadFile(path.Join(inputPath))
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", inputPath, err)
	}
	output := blackfriday.Run(f)

	if len(output) == 0 {
		return nil, fmt.Errorf("empty output from blackfriday")
	}
	doc, err := html.Parse(bytes.NewReader(output))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}
	return doc, nil
}

// GenerateHTMLs parses all translations and generates HTML output for all files.
// If strict is true, it will fail if any translation is missing, but
// when strict is false, base language will be substituted for any missing translations.
func GenerateHTMLs(baseLang string, targetLang []string, projectDir string, strict bool) error {
	slog.Debug("generating HTMLs",
		"baseLang", baseLang,
		"targetLang", targetLang,
		"projectDir", projectDir,
		"strict", strict,
	)

	// tx[lang][filename][key] = {translation string}
	tx := make(map[string]map[string]map[string]string)
	// set base translation as fallback
	dirTx, err := os.ReadDir(path.Join(projectDir, DefaultDirNameTranslations))
	if err != nil {
		return fmt.Errorf("read translations directory: %w", err)
	}

	// add existing translations from file
	for _, file := range dirTx {
		parts := strings.Split(file.Name(), ".")
		if len(parts) < 1 {
			return fmt.Errorf("invalid translation file name %q, expected format <lang>.<file>.json", file.Name())
		}
		// expects language code prefix (e.g. "en.README.json")
		lang := parts[0]
		name := parts[1]
		if !slices.Contains(targetLang, lang) && lang != baseLang {
			slog.Warn(
				"skipping translation file, not in target languages",
				"file", file.Name(),
				"baseLang", baseLang,
				"targetLangs", targetLang,
			)
			continue
		}

		f, err := os.ReadFile(path.Join(projectDir, DefaultDirNameTranslations, file.Name()))
		if err != nil {
			return fmt.Errorf("read translation file %q: %w", file.Name(), err)
		}
		fileTx := make(map[string]string)
		err = json.Unmarshal(f, &fileTx)
		if err != nil {
			return fmt.Errorf("unmarshal translation file %q: %w", file.Name(), err)
		}
		if len(fileTx) == 0 {
			return fmt.Errorf("translation file %q is empty or invalid", file.Name())
		}
		slog.Debug("translation file loaded",
			"file", file.Name(),
			"lang", lang,
			"textCount", len(fileTx),
		)
		if _, exists := tx[lang]; !exists {
			tx[lang] = make(map[string]map[string]string)
		}
		tx[lang][name] = fileTx
	}

	// Validate all translations are present.
	// If strict is true, it will fail if any translation is missing,
	// else, it will substitute missing translations from the base language
	for langName, langData := range tx {
		slog.Debug("processing translations", "language", langName)
		for fileName, fileData := range langData {
			slog.Debug("processing translation file", "language", langName, "file", fileName)
			for key, value := range fileData {
				// TODO: stuff
				if strings.TrimSpace(value) == "" {
					if strict {
						return fmt.Errorf(
							"[strict] missing translation %q in %q for language %q",
							key, fileName, langName,
						)
					}
					slog.Debug(
						"[no strict] empty translation value to be substituted with base language",
						"key", key,
						"file", fileName,
						"lang", langName,
						"baseLang", baseLang,
						"baseTxSub", tx[baseLang][fileName][key],
					)
					tx[langName][fileName][key] = tx[baseLang][fileName][key]
				}
			}
		}
	}

	// generate HTML file for every existing translation
	dirOut := path.Join(projectDir, DefaultDirNameOutput)
	err = os.MkdirAll(dirOut, DefaultFilePermissions)
	if err != nil {
		return fmt.Errorf("create output directory %v: %w", DefaultDirNameOutput, err)
	}

	// for each language, generate HTML files from templates and translations
	for lang, targetTx := range tx {
		slog.Debug("generating HTML files for language", "lang", lang)
		// for each translation file in the target language
		for txFileBasename, txData := range targetTx {
			slog.Debug("processing translation file", "lang", lang, "file", txFileBasename)
			// construct output file name
			outFile := fmt.Sprintf("%s.%s.html", lang, txFileBasename)
			outPath := path.Join(dirOut, outFile)
			fo, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("create output file %v: %w", outFile, err)
			}
			// load template file
			tmplFilename := fmt.Sprintf("%s.html.tmpl", txFileBasename)
			tmplPath := path.Join(projectDir, DefaultDirNameTemplates, tmplFilename)
			tmpl, err := template.ParseFiles(tmplPath)
			if err != nil {
				return fmt.Errorf("parse template %v: %w", tmplPath, err)
			}
			slog.Debug("generating HTML from template",
				"template", tmplFilename,
				"translation", txFileBasename,
			)
			// execute template with translation data
			err = tmpl.Execute(fo, txData)
			if err != nil {
				return fmt.Errorf("execute template %v: %w", txFileBasename, err)
			}
			err = fo.Close()
			if err != nil {
				return fmt.Errorf("close output file %v: %w", outFile, err)
			}
			slog.Info("generated HTML", "file", outFile)
		}
		slog.Info("generated HTML files for language",
			"lang", lang,
			"count", len(targetTx),
		)
	}
	return nil
}
