package illuminated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/russross/blackfriday/v2"
	"golang.org/x/net/html"
)

func Do() {
	input := path.Join("sample", "downloads.md")
	outHTML := path.Join("sample", "downloads.html.tmpl")
	outJSON := path.Join("sample", "en.json")

	var textToTranslate = make(map[string]string)
	doc, err := parse(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing input: %v\n", err)
		return
	}
	var counter int
	extractTemplate(doc, textToTranslate, &counter)
	for k, v := range textToTranslate {
		fmt.Printf("%s: %s\n", k, v)
	}
	fmt.Println("Extraction complete: \n", *doc)

	err = html.Render(os.Stdout, doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering HTML: %v\n", err)
	}

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

// extractTemplate extracts innerHTML strings into a map and
// replaces innerHTML with placeholders for internationalization.
func extractTemplate(n *html.Node, text map[string]string, counter *int) {
	if n.Type == html.TextNode {
		if len(strings.TrimSpace(n.Data)) > 0 {
			*counter++                               // increment field number...
			key := fmt.Sprintf("key_%02d", *counter) // to use as key for translation values,
			text[key] = n.Data                       // capture into map for translation file, and
			n.Data = fmt.Sprintf("{{ .%s }}", key)   // replace innerHTML with template placeholder
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractTemplate(c, text, counter)
	}
}

// parse converts markdown to HTML
func parse(inputPath string) (*html.Node, error) {
	f, err := os.ReadFile(path.Join(inputPath))
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", inputPath, err)
	}
	output := blackfriday.Run(f)

	fmt.Println(string(output))

	if len(output) == 0 {
		return nil, fmt.Errorf("empty output from blackfriday")
	}
	doc, err := html.Parse(bytes.NewReader(output))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}
	return doc, nil
}

// writeJSON writes a map[string]string to path
func writeJSON(path string, data map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file %q: %w", path, err)
	}
	defer file.Close()

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	_, err = file.Write(b)
	if err != nil {
		return fmt.Errorf("write json to file: %w", err)
	}
	return nil
}

func writePDF(sourcePath, outPath string) error {
	cmd := exec.Command("pandoc", sourcePath, "-o", outPath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("execute pandoc command: %w", err)
	}
	return nil
}

func writeHTML(path string, doc *html.Node) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file %q: %w", path, err)
	}
	defer file.Close()
	err = html.Render(file, doc)
	if err != nil {
		return fmt.Errorf("render html to file: %w", err)
	}
	return nil
}
