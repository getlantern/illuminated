//go:build integration

package translators

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoogle(t *testing.T) {
	ctx := context.Background()
	g, err := NewGoogleTranslator(ctx)
	require.NoError(t, err)
	defer g.Close(ctx)
	langs, err := g.SupportedLanguages(ctx, "en")
	require.NoError(t, err)
	require.NotEmpty(t, langs)
	for _, lang := range langs {
		t.Logf("Supported language: %s", lang)
	}
}

func TestGoogleHTML(t *testing.T) {
	ctx := context.Background()
	g, err := NewGoogleTranslator(ctx)
	require.NoError(t, err)
	defer g.Close(ctx)

	// Example HTML content to translate
	htmlContent := "<p>Hello, world!</p>"
	targetLang := "es" // Spanish

	// Translate
	translations, err := g.Translate(ctx, targetLang, []string{htmlContent})
	require.NoError(t, err)
	require.Len(t, translations, 1)
	t.Logf("Translated HTML: %s", translations[0])
}
