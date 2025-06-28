package illuminated

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

var testProjectDir = path.Join("test-project")

func TestMarkdownToRawHTML(t *testing.T) {
	input := "test.md"
	err := os.WriteFile(input, []byte("# Hello World"), 0o644)
	require.NoError(t, err)
	defer os.Remove(input)

	html, err := markdownToRawHTML(input)
	require.NoError(t, err)
	require.Contains(t, html, "<h1>Hello World</h1>")
}

func TestMarkdownToHTML(t *testing.T) {
	input := "test.md"
	output := "test.html"
	err := os.WriteFile(input, []byte("# Hello World"), 0o644)
	require.NoError(t, err)
	defer os.Remove(input)
	defer os.Remove(output)

	err = MarkdownToHTML(input, output)
	require.NoError(t, err)

	content, err := os.ReadFile(output)
	require.NoError(t, err)
	require.Contains(t, string(content), "<h1>Hello World</h1>")
}
