// package illuminated converts markdown into corresponding templates and translations,
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
//   - <DirTranslations>/<lang>.json       (translation strings)
//   - <DirTranslations>/<file>.html.tmpl  (go template)
func Process(input string, projectDir string) error {
	var counter int
	var translationStrings = make(map[string]string)
	doc, err := parseHTML(input)
	if err != nil {
		return fmt.Errorf("parse input: %v", err)
	}
	extractInnerHTML(doc, translationStrings, &counter)
	for k, v := range translationStrings {
		slog.Debug("extracted", "key", k, "value", v, "file", input)
	}
	baseName := strings.TrimSuffix(path.Base(input), path.Ext(input))

	// json
	dirTranslations := path.Join(projectDir, DefaultDirNameTranslations)
	err = os.MkdirAll(dirTranslations, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create directory %q: %v", dirTranslations, err)
	}
	jsonOut := path.Join(dirTranslations, fmt.Sprintf("%s.%s.json", DefaultConfig.Base, baseName))
	err = writeJSON(jsonOut, translationStrings)
	if err != nil {
		return fmt.Errorf("write %v: %v", jsonOut, err)
	}
	slog.Debug("translation strings written", "file", jsonOut)

	// template
	dirTemplates := path.Join(projectDir, DefaultDirNameTemplates)
	err = os.MkdirAll(dirTemplates, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create directory %v: %v", dirTemplates, err)
	}
	tmplOut := path.Join(dirTemplates, fmt.Sprintf("%s.html.tmpl", baseName))
	err = writeHTML(tmplOut, doc)
	if err != nil {
		return fmt.Errorf("write HTML template: %v", err)
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

// parseHTML converts markdown file to HTML object
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
		return fmt.Errorf("read translations directory: %v", err)
	}
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
				return fmt.Errorf("read base translation file %v: %v", baseTxFile, err)
			}
			err = json.Unmarshal(f, &fileTx)
			if err != nil {
				return fmt.Errorf("unmarshal base translation file %v: %v", baseTxFile, err)
			}
			baseTx[baseLang] = fileTx
		}
	}

	// generate translated HTMLs
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
					return fmt.Errorf("read target translation file %v: %v", targetTxFile, err)
				}
				err = json.Unmarshal(f, &fileTx)
				if err != nil {
					return fmt.Errorf("unmarshal target translation file %v: %v", targetTxFile, err)
				}
				targetTx[file.Name()] = fileTx
			}
		}

		// validate and/or substitute depending on strict mode
		for filename, file := range targetTx {
			for k, v := range file {
				if v == "" {
					if strict {
						return fmt.Errorf("missing translation %q in %q", k, file)
					} else {
						slog.Warn("missing translation, substituting base lang string",
							"key", k,
							"file", file,
							"baseLang", baseLang,
							"targetLang", lang,
							"baseTx", baseTx[baseLang][k],
						)
					}
					targetTx[filename][k] = baseTx[filename][k]
				}
			}
		}

		// generate HTML files
		dirOut := path.Join(projectDir, DefaultDirNameOutput)
		err = os.MkdirAll(dirOut, os.ModePerm)
		if err != nil {
			return fmt.Errorf("create output directory %v: %v", DefaultDirNameOutput, err)
		}
		for _, file := range dirTx {
			outFile := fmt.Sprintf("%s.%s.html", lang, strings.TrimPrefix(file.Name(), lang+"."))
			fo, err := os.Create(path.Join(dirOut, outFile))
			if err != nil {
				return fmt.Errorf("create output file %v: %v", outFile, err)
			}
			defer fo.Close()
			tmplFilename := strings.TrimPrefix(file.Name(), lang+".")
			tmplFilename = strings.TrimSuffix(tmplFilename, path.Ext(tmplFilename)) + ".html.tmpl"
			tmpl, err := template.ParseFiles(path.Join(
				projectDir,
				DefaultDirNameTemplates,
				tmplFilename,
			))
			if err != nil {
				return fmt.Errorf("parse template %v: %v", tmplFilename, err)
			}
			err = tmpl.Execute(fo, targetTx[file.Name()])
			if err != nil {
				return fmt.Errorf("execute template %v: %v", file.Name(), err)
			}
			slog.Info("generated HTML", "file", outFile)
		}
	}
	return nil
}
