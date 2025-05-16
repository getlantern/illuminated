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

	// set base translation as fallback
	baseTx := make(map[string]map[string]string)
	dirTx, err := os.ReadDir(path.Join(projectDir, DefaultDirNameTranslations))
	if err != nil {
		return fmt.Errorf("read translations directory: %w", err)
	}

	// add the base language translation strings to the baseTx map
	for _, f := range dirTx {
		if strings.HasPrefix(f.Name(), baseLang+".") {
			fileTx := make(map[string]string)
			baseTxFile := path.Join(
				projectDir,
				DefaultDirNameTranslations,
				f.Name(),
			)
			slog.Debug("extracting base language translations strings", "file", baseTxFile)
			f, err := os.ReadFile(baseTxFile)
			if err != nil {
				return fmt.Errorf("read base translation file %v: %w", baseTxFile, err)
			}
			err = json.Unmarshal(f, &fileTx)
			if err != nil {
				return fmt.Errorf("unmarshal base translation file %v: %w", baseTxFile, err)
			}
			baseTx[baseLang] = fileTx
		}
	}

	// generate translated HTMLs for every target language
	for _, lang := range targetLang {
		// populate target translation map
		// lang.file.key
		targetTx := make(map[string]map[string]string)
		for _, file := range dirTx {
			if strings.HasPrefix(file.Name(), lang+".") {
				fileTx := make(map[string]string)
				targetTxFile := path.Join(
					projectDir,
					DefaultDirNameTranslations,
					file.Name(),
				)
				slog.Debug("reading target translation file", "file", targetTxFile)
				f, err := os.ReadFile(targetTxFile)
				if err != nil {
					return fmt.Errorf("read target translation file %v: %w", targetTxFile, err)
				}
				err = json.Unmarshal(f, &fileTx)
				if err != nil {
					return fmt.Errorf("unmarshal target translation file %v: %w", targetTxFile, err)
				}
				targetTx[file.Name()] = fileTx
			}
		}
		// validate and/or substitute (depending on strict mode)
		for filename, file := range targetTx {
			for k, v := range file {
				if strings.TrimSpace(v) == "" {
					if strict {
						return fmt.Errorf("missing translation %q in %q", k, filename)
					} else {
						slog.Warn("missing translation, substituting base lang string",
							"key", k,
							"file", filename,
							"baseLang", baseLang,
							"targetLang", lang,
							"baseTx", baseTx[baseLang][k],
						)
					}
					targetTx[filename][k] = baseTx[baseLang][k]
				}
			}
		}

		// generate HTML file for every existing translation
		dirOut := path.Join(projectDir, DefaultDirNameOutput)
		err = os.MkdirAll(dirOut, DefaultFilePermissions)
		if err != nil {
			return fmt.Errorf("create output directory %v: %w", DefaultDirNameOutput, err)
		}
		// TODO: throw an error if there is a template but not translation or vice versa.
		for _, tx := range dirTx {
			if !strings.HasPrefix(tx.Name(), lang+".") {
				continue // skip non-matching languages
			}
			outFile := fmt.Sprintf("%s.%s.html",
				lang, strings.TrimPrefix(strings.TrimSuffix(tx.Name(), ".json"), lang+"."),
			)
			outPath := path.Join(dirOut, outFile)
			fo, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("create output file %v: %w", outFile, err)
			}
			tmplFilename := strings.TrimPrefix(tx.Name(), lang+".")
			tmplFilename = strings.TrimSuffix(tmplFilename, path.Ext(tmplFilename)) + ".html.tmpl"
			tmpl, err := template.ParseFiles(path.Join(
				projectDir,
				DefaultDirNameTemplates,
				tmplFilename,
			))
			slog.Debug("generating HTML from template",
				"template", tmplFilename,
				"translation", tx.Name(),
			)
			if err != nil {
				return fmt.Errorf("parse template %v: %w", tmplFilename, err)
			}
			err = tmpl.Execute(fo, targetTx[tx.Name()])
			if err != nil {
				return fmt.Errorf("execute template %v: %w", tx.Name(), err)
			}
			fo.Close()
			slog.Info("generated HTML", "file", outFile)
		}
	}
	return nil
}
