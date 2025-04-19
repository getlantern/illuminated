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
	slog.Debug("calling pandoc to write write from HTML", "source", sourcePath, "out", outPath)
	err := formatBreaks(sourcePath)
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
		// "--standalone",
		// "--resource-path", resourcePath,
		// "--pdf-engine", "weasyprint",
		"--pdf-engine", "pdflatex",
		// "--embed-resources",
		sourcePath, "-o", outPath,
	)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("execute pandoc command: %w", err)
	}
	return nil
}

func formatBreaks(filepathHTML string) error {
	htmlContent, err := os.ReadFile(filepathHTML)
	if err != nil {
		return fmt.Errorf("read HTML file: %v", err)
	}

	// Add a page break before every <h1> tag
	modifiedHTML := strings.ReplaceAll(
		string(htmlContent),
		"<h1>",
		"<br><h1>",
		// FEATURE: add proper page break before each chapter
		// none of these work, gah
		// `<div style="display:block; clear:both; page-break-before:always;"></div><h1>`,
		// `<p>\newpage</p><h1>`,
		// `<div class="page-break"></div><h1>`,
		// "\n\n\\newpage\n\n<h1>",
		// `<b>\newpage</b><h1>`,
		// `<h1 style="page-break-before: always;">`,
	)

	// Write the modified HTML back to the file
	err = os.WriteFile(filepathHTML, []byte(modifiedHTML), 0644)
	if err != nil {
		return fmt.Errorf("write modified HTML: %v", err)
	}

	return nil
}

func JoinHTML(language string, projectDir string, name string) (string, error) {
	outputDir := path.Join(projectDir, DefaultDirNameOutput)
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return "", fmt.Errorf("read output directory: %v", err)
	}
	joinedFilePath := path.Join(outputDir, fmt.Sprintf("%s.html", name))
	joinedFile, err := os.Create(joinedFilePath)
	if err != nil {
		return "", fmt.Errorf("create consolidated file: %v", err)
	}
	defer joinedFile.Close()

	var combinedBody strings.Builder
	combinedBody.WriteString("<html>\n<head></head>\n<body>\n")

	for _, file := range files {
		if file.IsDir() {
			slog.Warn("skipping unexpected directory in output dir", "name", file.Name())
			continue
		}
		if !strings.HasPrefix(file.Name(), language+".") || !strings.HasSuffix(file.Name(), ".html") {
			slog.Debug("skipping file", "name", file.Name())
			continue
		}

		filePath := path.Join(outputDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read file %v: %v", file.Name(), err)
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
			return "", fmt.Errorf("delete file %v: %v", file.Name(), err)
		}
	}

	combinedBody.WriteString("</body>\n</html>")

	// Write the combined HTML to the output file
	_, err = joinedFile.WriteString(combinedBody.String())
	if err != nil {
		return "", fmt.Errorf("write to consolidated file: %v", err)
	}

	return joinedFilePath, nil
}
