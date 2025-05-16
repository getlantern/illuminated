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
func WritePDF(sourcePath, outPath string, resourcePath string) error {
	slog.Debug("calling pandoc to write from HTML", "source", sourcePath, "out", outPath)
	// first verify that pandoc is installed
	_, err := exec.LookPath("pandoc")
	if err != nil {
		return fmt.Errorf("pandoc not found in PATH, install and try again: %w", err)
	}

	err = formatBreaks(sourcePath)
	if err != nil {
		return fmt.Errorf("format breaks in HTML: %w", err)
	}

	title := strings.TrimSuffix(sourcePath, ".html")

	if resourcePath == "" {
		resourcePath = "."
	}

	cmd := exec.Command(
		"pandoc",
		"--metadata", fmt.Sprintf("title=%s", path.Base(title)),
		"--metadata", fmt.Sprintf("date=%s", time.Now().Format("2006-01-02")),
		"--toc",
		"--pdf-engine", "pdflatex",
		sourcePath, "-o", outPath,
	)

	err = cmd.Run()
	if err != nil {
		if strings.Contains(err.Error(), "47") {
			return fmt.Errorf(
				"pandoc: pdf engine (such as latex) not found or invalid; install and try again: %w",
				err,
			)
		}
		return fmt.Errorf("pandoc: %w", err)
	}
	return nil
}

// formatBreaks adds a break before each <h1> tag in the HTML file.
func formatBreaks(filepathHTML string) error {
	htmlContent, err := os.ReadFile(filepathHTML)
	if err != nil {
		return fmt.Errorf("read HTML file: %w", err)
	}

	// Add a page break before every <h1> tag
	modifiedHTML := strings.ReplaceAll(
		string(htmlContent),
		"<h1>",
		"<br><h1>", // just use a break for now :(
		// TODO: format in a way that LaTeX respects as full page break.
		// add proper page break before each chapter.
		// Investigate why I am unable to inject a page break into HTML
		// that pancdoc will respect as a LaTeX page break.
		//
		// Graveyard of attempts:
		// `<div style="display:block; clear:both; page-break-before:always;"></div><h1>`,
		// `<p>\newpage</p><h1>`,
		// `<div class="page-break"></div><h1>`,
		// "\n\n\\newpage\n\n<h1>",
		// `<b>\newpage</b><h1>`,
		// `<h1 style="page-break-before: always;">`,
	)

	// Write the modified HTML back to the file
	err = os.WriteFile(filepathHTML, []byte(modifiedHTML), 0o644)
	if err != nil {
		return fmt.Errorf("write modified HTML: %w", err)
	}

	return nil
}

// JoinHTML combines all HTML files for a given language (denoted by prefix)
// into a single HTML file. This may be an intermediary step before PDF generation.
func JoinHTML(language string, projectDir string, name string) (string, error) {
	outputDir := path.Join(projectDir, DefaultDirNameOutput)
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return "", fmt.Errorf("read output directory: %w", err)
	}
	joinedFilePath := path.Join(outputDir, fmt.Sprintf("%s.%s.html", language, name))
	joinedFile, err := os.Create(joinedFilePath)
	if err != nil {
		return "", fmt.Errorf("create consolidated file: %w", err)
	}
	defer joinedFile.Close()

	var combinedBody strings.Builder
	combinedBody.WriteString("<html>\n<head></head>\n<body>\n")

	for _, file := range files {
		if file.IsDir() {
			slog.Warn("skipping unexpected directory in output dir", "name", file.Name())
			continue
		}
		if !strings.HasPrefix(file.Name(), language+".") {
			slog.Debug("skipping on language mismatch",
				"name", file.Name(),
				"lang", language,
			)
			continue
		}
		if !strings.HasSuffix(file.Name(), ".html") {
			slog.Debug("skipping non-HTML file", "name", file.Name())
			continue
		}

		filePath := path.Join(outputDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read file %v: %w", file.Name(), err)
		}

		bodyStart := strings.Index(string(content), "<body>")
		bodyEnd := strings.Index(string(content), "</body>")
		if bodyStart == -1 || bodyEnd == -1 {
			slog.Warn("skipping file with no <body> tag", "name", file.Name())
			continue
		}
		bodyContent := string(content)[bodyStart+len("<body>") : bodyEnd]
		combinedBody.WriteString(bodyContent)
		combinedBody.WriteString("\n")

		err = os.Remove(filePath)
		if err != nil {
			return "", fmt.Errorf("delete file %v: %w", file.Name(), err)
		}
	}

	combinedBody.WriteString("</body>\n</html>")

	// Write the combined HTML to the output file
	_, err = joinedFile.WriteString(combinedBody.String())
	if err != nil {
		return "", fmt.Errorf("write to consolidated file: %w", err)
	}

	return joinedFilePath, nil
}
