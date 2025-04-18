package illuminated

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/net/html"
)

// writeJSON writes a map[string]string to path as JSON.
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

// writeHTML writes an HTML document to file at path.
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

// WritePDF calls pandoc to output a PDF from a source file (HTML expected).
func WritePDF(sourcePath, outPath string) error {
	cmd := exec.Command("pandoc", "--toc", sourcePath, "-o", outPath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("execute pandoc command: %w", err)
	}
	return nil
}
