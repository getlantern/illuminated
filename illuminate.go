package illuminated

import (
	"bytes"
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
	BaseLang        = "en"           // ISO 639-1 language code
	DirStaging      = "staging"      // copies of source and intermediate files
	DirOutput       = "output"       // final output (typically PDF)
	DirTranslations = "translations" // translation files for internationalization
	DirTemplates    = "templates"    // template to recreate localized copies
)

func Do() {
	input := path.Join("sample", "downloads.md")
	outHTML := path.Join("sample", "downloads.html.tmpl")
	outJSON := path.Join("sample", "en.json")

	doc, textToTranslate, err := Process(input)

	// err = html.Render(os.Stdout, doc)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error rendering HTML: %v\n", err)
	// }

	if err := writeJSON(outJSON, textToTranslate); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON: %v\n", err)
	}
	if err := writeHTML(outHTML, doc); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing HTML: %v\n", err)
	}

	tmpl, err := template.ParseFiles(outHTML)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		return
	}

	outFile, err := os.Create("sample/downloads.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		return
	}
	defer outFile.Close()

	err = tmpl.Execute(outFile, textToTranslate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing template to file: %v\n", err)
	}

	err = writePDF("sample/downloads.html", "sample/downloads.pdf")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing PDF: %v\n", err)
	}
	fmt.Println("Success!")

}

// Process an input markdown file path into parts:
//   - HTML document
//   - map of translation strings
func Process(input string) (*html.Node, map[string]string, error) {
	var counter int
	var translationStrings = make(map[string]string)
	doc, err := parse(input)
	if err != nil {
		return nil, nil, fmt.Errorf("parse input: %v", err)
	}
	extract(doc, translationStrings, &counter)
	for k, v := range translationStrings {
		slog.Debug("extracted", "key", k, "value", v, "file", input)
	}
	baseName := strings.TrimSuffix(path.Base(input), path.Ext(input))

	jsonOut := path.Join(DirTranslations, fmt.Sprintf("%s.json", baseName))
	err = writeJSON(jsonOut, translationStrings)
	if err != nil {
		return nil, nil, fmt.Errorf("write %q: %v", jsonOut, err)
	}
	slog.Debug("translation strings written", "file", jsonOut)

	htmlOut := path.Join(DirStaging, fmt.Sprintf("%s.html", baseName))
	err = writeHTML(htmlOut, doc)
	if err != nil {
		return nil, nil, fmt.Errorf("write HTML template: %v", err)
	}
	slog.Debug("HTML template written", "file", htmlOut)

	tmplOut := fmt.Sprintf("%s.html.tmpl", baseName)
	tmpl, err = template.ParseFiles(htmlOut)
	if err != nil {
		return nil, nil, fmt.Errorf("parse template: %v", err)
	}
	outFile, err = os.Create(path.Join(DirTemplates, tmplOut))
	if err != nil {
		return nil, nil, fmt.Errorf("create template %q: %v", tmplOut, err)

	return doc, translationStrings, nil
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

// parse converts markdown to HTML
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
