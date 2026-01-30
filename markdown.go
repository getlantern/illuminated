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
	output := blackfriday.Run(fillSpacingMD(f))

	return string(output), nil
}

// fillSpacingMD adds spacing in markdown data where needed to ensure proper HTML rendering.
// - after headers
// - before lists
func fillSpacingMD(data []byte) []byte {
	// Add blank line after headers (## Header) if followed by non-blank content
	// Match: header line followed immediately by non-blank, non-header line
	headerPattern := regexp.MustCompile(`(?m)(^#{1,6}\s+.+)\n([^\n#])`)
	data = headerPattern.ReplaceAll(data, []byte("$1\n\n$2"))

	// Add blank line before unordered lists (- item) if not already present
	// Match: non-blank line followed immediately by list item
	unorderedListPattern := regexp.MustCompile(`(?m)([^\n])\n(^-\s+)`)
	data = unorderedListPattern.ReplaceAll(data, []byte("$1\n\n$2"))

	// Add blank line before ordered lists (1. item) if not already present
	// Match: non-blank line followed immediately by numbered list item
	orderedListPattern := regexp.MustCompile(`(?m)([^\n])\n(^\d+\.\s+)`)
	data = orderedListPattern.ReplaceAll(data, []byte("$1\n\n$2"))

	return data
}

// MarkdownToHTML reads markdown from inputPath and writes HTML to outputPath.
func MarkdownToHTML(inputPath string, outputPath string) error {
	doc, err := markdownToRawHTML(inputPath)
	if err != nil {
		return err
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file %q: %w", outputPath, err)
	}
	defer f.Close()

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
	_, err = f.WriteString(wrapped)
	if err != nil {
		return fmt.Errorf("write to output file %q: %w", outputPath, err)
	}
	slog.Debug("HTML output generated",
		"input", inputPath,
		"output", outputPath,
	)

	return nil
}
