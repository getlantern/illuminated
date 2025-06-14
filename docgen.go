package illuminated

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/russross/blackfriday/v2"
)

var DefaultDirNameHTML = "html"

// markdownToRawHTML reads a file from inputPath, returning an HTML string.
func markdownToRawHTML(inputPath string) (string, error) {
	f, err := os.ReadFile(path.Join(inputPath))
	if err != nil {
		return "", fmt.Errorf("read file %q: %w", inputPath, err)
	}
	output := blackfriday.Run(f)

	return string(output), nil
}

// markdownToHTML reads markdown from inputPath and writes HTML to outputPath.
func markdownToHTML(inputPath string, outputPath string) error {
	doc, err := markdownToRawHTML(inputPath)
	if err != nil {
		return err
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file %q: %w", outputPath, err)
	}
	defer f.Close()

	_, err = f.WriteString(doc)
	if err != nil {
		return fmt.Errorf("write to output file %q: %w", outputPath, err)
	}
	slog.Info("HTML output generated",
		"input", inputPath,
		"output", outputPath,
	)

	return nil
}

// func generateTranslationHTMLs(baseLangFilepath string, targetLangs []string) error {
//     for _, lang := range targetLangs {
//         if lang == baseLang {
//             continue
//         }
//         outputFile := path.Join(DefaultDirNameHTML, fmt.Sprintf("%s.%s.html", lang, path.Base(baseLangFilepath)))
//     }
//     return nil
// }
