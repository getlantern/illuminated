package illuminated

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/russross/blackfriday/v2"
)

var DefaultDirNameHTML = "html"

// markdownToRawHTML reads a file from inputPath, returning an HTML string.
func markdownToRawHTML(inputPath string) (string, error) {
	f, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("read file %q: %w", inputPath, err)
	}
	output := blackfriday.Run(fixMarkdownSpacing(f))
	return string(output), nil
}

// fixMarkdownSpacing adds blank lines after headers and before lists to ensure proper HTML rendering
func fixMarkdownSpacing(data []byte) []byte {
	// Blank line after headers
	data = regexp.MustCompile(`(?m)(^#{1,6}\s+.+)\n([^\n#])`).ReplaceAll(data, []byte("$1\n\n$2"))
	// Blank line before unordered lists
	data = regexp.MustCompile(`(?m)([^\n])\n(^-\s+)`).ReplaceAll(data, []byte("$1\n\n$2"))
	// Blank line before ordered lists
	data = regexp.MustCompile(`(?m)([^\n])\n(^\d+\.\s+)`).ReplaceAll(data, []byte("$1\n\n$2"))
	return data
}

// MarkdownToHTML reads markdown from inputPath and writes HTML to outputPath.
func MarkdownToHTML(inputPath string, outputPath string) error {
	doc, err := markdownToRawHTML(inputPath)
	if err != nil {
		return err
	}

	wrapped := fmt.Sprintf(
		`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body>
%s
</body>
</html>`, doc)

	if err := os.WriteFile(outputPath, []byte(wrapped), DefaultFilePermissions); err != nil {
		return fmt.Errorf("write output file %q: %w", outputPath, err)
	}

	slog.Debug("HTML output generated",
		"input", inputPath,
		"output", outputPath,
	)
	return nil
}
