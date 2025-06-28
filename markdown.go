package illuminated

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/russross/blackfriday/v2"
)

var DefaultDirNameHTML = "html"

// markdownToRawHTML reads a file from inputPath, returning an HTML string.
func markdownToRawHTML(inputPath string) (string, error) {
	f, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("read file %q: %w", inputPath, err)
	}
	output := blackfriday.Run(f)

	return string(output), nil
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
