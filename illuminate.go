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

var (
	BaseLang = "en" // ISO 639-1 language code
	// consider using an optional config file .illuminatedrc (maybe an init commmand with prompt)
	DirStaging      = "staging"      // copies of source and intermediate files
	DirOutput       = "output"       // final output (typically PDF)
	DirTranslations = "translations" // translation files for internationalization
	DirTemplates    = "templates"    // template to recreate localized copies
)

// Process an input markdown file into parts:
//   - <DirTranslations>/<lang>.json       (translation strings)
//   - <DirTranslations>/<file>.html.tmpl  (go template)
func Process(input string) error {
	var counter int
	var translationStrings = make(map[string]string)
	doc, err := parse(input)
	if err != nil {
		return fmt.Errorf("parse input: %v", err)
	}
	extract(doc, translationStrings, &counter)
	for k, v := range translationStrings {
		slog.Debug("extracted", "key", k, "value", v, "file", input)
	}
	baseName := strings.TrimSuffix(path.Base(input), path.Ext(input))

	// json
	err = os.MkdirAll(DirTranslations, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create directory %q: %v", DirTranslations, err)
	}
	jsonOut := path.Join(DirTranslations, fmt.Sprintf("%s.%s.json", BaseLang, baseName))
	err = writeJSON(jsonOut, translationStrings)
	if err != nil {
		return fmt.Errorf("write %v: %v", jsonOut, err)
	}
	slog.Info("translation strings written", "file", jsonOut)

	// template
	err = os.MkdirAll(DirTemplates, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create directory %v: %v", DirTemplates, err)
	}
	tmplOut := path.Join(DirTemplates, fmt.Sprintf("%s.html.tmpl", baseName))
	err = writeHTML(tmplOut, doc)
	if err != nil {
		return fmt.Errorf("write HTML template: %v", err)
	}
	slog.Info("HTML template written", "file", tmplOut)
	return nil
}

// Generate combines <DirTranslation>/<lang>.json and <DirTemplates>/html.tmpl to make:
//   - <DirOutput>/<lang>.<name>.html
//   - <DirOutput>/<lang>.<name>.pdf
func Generate(name string, langCode string) error {
	htmlOut := fmt.Sprintf("%s.%s.html", langCode, name)
	targetTemplate := path.Join(DirTemplates, fmt.Sprintf("%s.html.tmpl", name))
	targetTranslations := path.Join(DirTranslations, fmt.Sprintf("%s.%s.json", langCode, name))
	f, err := os.ReadFile(targetTranslations)
	if err != nil {
		return fmt.Errorf("read translation file %v: %v", targetTranslations, err)
	}
	var translations map[string]string
	err = json.Unmarshal(f, &translations)
	if err != nil {
		return fmt.Errorf("unmarshal translation file %v: %v", targetTranslations, err)
	}

	// TODO check for missing translations
	fallbackNecessary := false
	for _, v := range translations {
		if v == "" {
			fallbackNecessary = true
		}
	}
	if fallbackNecessary {
		slog.Warn("translation missing", "file", targetTranslations)
		// TODO fallback to base lang selectively
	}

	// generate html
	tmpl, err := template.ParseFiles(targetTemplate)
	if err != nil {
		return fmt.Errorf("parse template %v: %v", targetTemplate, err)
	}
	err = os.MkdirAll(DirOutput, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create directory %v: %v", DirOutput, err)
	}
	outFile, err := os.Create(path.Join(DirOutput, htmlOut))
	if err != nil {
		return fmt.Errorf("create template %v: %v", htmlOut, err)
	}
	defer outFile.Close()
	err = tmpl.Execute(outFile, translations)
	if err != nil {
		return fmt.Errorf("execute template: %v", err)
	}

	// generate pdf
	pdfOut := fmt.Sprintf("%s.%s.pdf", langCode, name)
	err = writePDF(path.Join(DirOutput, htmlOut), pdfOut)
	if err != nil {
		return fmt.Errorf("generate PDF: %v", err)
	}
	slog.Info("generated", "pdf", pdfOut, "html", htmlOut)

	return nil
}

// extract extracts innerHTML strings into a map and
// replaces innerHTML with placeholders for internationalization.
func extract(n *html.Node, text map[string]string, counter *int) {
	if n.Type == html.TextNode {
		if len(strings.TrimSpace(n.Data)) > 0 {
			*counter++                               // increment field number...
			key := fmt.Sprintf("key_%02d", *counter) // to use as key for translation values,
			text[key] = n.Data                       // capture into map for translation file, and
			n.Data = fmt.Sprintf("{{ .%s }}", key)   // replace innerHTML with template placeholder
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extract(c, text, counter)
	}
}

// parse converts markdown file to HTML object
func parse(inputPath string) (*html.Node, error) {
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
