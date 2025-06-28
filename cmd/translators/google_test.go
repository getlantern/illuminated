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
