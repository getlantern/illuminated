package illuminated

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

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
	slog.Debug("calling pandoc to write write from HTML", "source", sourcePath, "out", outPath)

	err := formatBreaks(sourcePath)
	if err != nil {
		return fmt.Errorf("format breaks in HTML: %w", err)
	}

	title := strings.TrimSuffix(sourcePath, ".html")
	cmd := exec.Command(
		"pandoc",
		"--metadata", fmt.Sprintf("title=%s", path.Base(title)),
		"--metadata", fmt.Sprintf("date=%s", time.Now().Format("2006-01-02")),
		"--toc",
		sourcePath, "-o", outPath,
	)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("execute pandoc command: %w", err)
	}
	return nil
}

func formatBreaks(filepathHTML string) error {
	// Read the HTML file
	htmlContent, err := os.ReadFile(filepathHTML)
	if err != nil {
		return fmt.Errorf("read HTML file: %v", err)
	}

	// Add a page break before every <h1> tag
	modifiedHTML := strings.ReplaceAll(
		string(htmlContent),
		"<h1>",
		`<br><br><br><br><br><h1>`,
		// FEATURE: add proper page break before each chapter
		// `<div style="display:block; clear:both; page-break-before:always;"></div><h1>`,
	)

	// Write the modified HTML back to the file
	err = os.WriteFile(filepathHTML, []byte(modifiedHTML), 0644)
	if err != nil {
		return fmt.Errorf("write modified HTML: %v", err)
	}

	return nil
}
