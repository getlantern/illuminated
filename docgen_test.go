package illuminated

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/getlantern/illuminated/translators"
	"github.com/stretchr/testify/require"
)

func TestBar(t *testing.T) {
	testDir := path.Join("fake")
	err := os.MkdirAll(testDir, DefaultFilePermissions)
	require.NoError(t, err)

	exampleDir := path.Join("example")
	err = MarkdownToHTML(path.Join(exampleDir, "downloads.md"), path.Join(testDir, "downloads.html"))
	require.NoError(t, err)

	err = os.RemoveAll(testProjectDir)
	require.NoError(t, err)

	ctx := context.Background()
	g, err := translators.NewGoogleTranslator(ctx)
	require.NoError(t, err)
	rawHTML, err := markdownToRawHTML(path.Join(exampleDir, "downloads.md"))
	require.NoError(t, err)
	result, err := g.Translate(ctx, "es", []string{rawHTML})
	require.NoError(t, err)
	require.Len(t, result, 1)
	t.Logf("Translated HTML:\n%s", result[0])
}
