package translators

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	gtx "cloud.google.com/go/translate"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

const maxTranslateBytes = 25000 // Google API limit ~30KB, use 25KB for safety

// googleTranslator implements the Translator interface using Google Translate API.
type googleTranslator struct {
	Client *gtx.Client
}

// NewGoogleTranslator returns a new Google Translate client.
func NewGoogleTranslator(ctx context.Context) (*googleTranslator, error) {
	keyName := "GOOGLE_API_KEY"
	key, ok := os.LookupEnv(keyName)
	if !ok {
		return nil, fmt.Errorf("%s not found in context", keyName)
	}
	client, err := gtx.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		return nil, fmt.Errorf("create Google Translate client: %w", err)
	}
	return &googleTranslator{Client: client}, nil
}

// SupportedLanguages returns a list of supported target languages for the given base language.
func (g *googleTranslator) SupportedLanguages(ctx context.Context, baseLang string) ([]string, error) {
	tag := language.Make(baseLang)
	slog.Debug("making tag from base lang", "baseLang", baseLang, "tag", tag)
	langs, err := g.Client.SupportedLanguages(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("get supported languages: %w", err)
	}
	var langCodes []string
	for _, lang := range langs {
		langCodes = append(langCodes, lang.Tag.String())
	}
	return langCodes, nil
}

// Translate translates the given texts into the target language.
func (g *googleTranslator) Translate(
	ctx context.Context,
	targetLang string,
	texts []string,
) ([]string, error) {
	if len(texts) != 1 {
		return nil, fmt.Errorf("expected 1 text, got %d", len(texts))
	}

	text := texts[0]
	if len(text) > maxTranslateBytes*10 {
		return nil, fmt.Errorf("content too large (%d bytes, max %d)", len(text), maxTranslateBytes*10)
	}

	tag := language.Make(targetLang)
	slog.Debug("making language tag for target lang", "tag", tag)
	opts := &gtx.Options{Format: gtx.HTML}

	// Try full translation first
	t, err := g.Client.Translate(ctx, []string{text}, tag, opts)
	if err == nil {
		for i, translation := range t {
			slog.Debug("translated text",
				"index", i,
				"source", translation.Source,
				"target", translation.Model,
				"text", translation.Text,
			)
		}
		return []string{t[0].Text}, nil
	}

	// On error, try chunked translation if content is large enough
	if len(text) < 2000 {
		return nil, fmt.Errorf("translate: %w", err)
	}

	slog.Info("retrying with chunked translation", "size", len(text))
	return g.translateChunked(ctx, tag, opts, text)
}

func (g *googleTranslator) translateChunked(ctx context.Context, tag language.Tag, opts *gtx.Options, html string) ([]string, error) {
	chunks := splitByH2(html)
	if len(chunks) == 1 {
		return nil, fmt.Errorf("cannot chunk content (no <h2> sections found)")
	}

	var result strings.Builder
	for i, chunk := range chunks {
		if len(chunk) > maxTranslateBytes {
			return nil, fmt.Errorf("chunk %d too large (%d bytes, max %d)", i, len(chunk), maxTranslateBytes)
		}
		t, err := g.Client.Translate(ctx, []string{chunk}, tag, opts)
		if err != nil {
			return nil, fmt.Errorf("translate chunk %d: %w", i, err)
		}
		for j, translation := range t {
			slog.Debug("translated text",
				"index", j,
				"source", translation.Source,
				"target", translation.Model,
				"text", translation.Text,
			)
		}
		result.WriteString(t[0].Text)
	}
	return []string{result.String()}, nil
}

func splitByH2(html string) []string {
	parts := regexp.MustCompile(`(?i)(<h2>)`).Split(html, -1)
	if len(parts) <= 1 {
		return []string{html}
	}

	var chunks []string
	if parts[0] != "" {
		chunks = append(chunks, parts[0])
	}
	for i := 1; i < len(parts); i++ {
		chunks = append(chunks, "<h2>"+parts[i])
	}
	return chunks
}

func (g *googleTranslator) Close(ctx context.Context) {
	if g == nil {
		slog.Debug("translator client 'Google Translate' is already nil, nothing to close")
		return
	}
	err := g.Client.Close()
	if err != nil {
		slog.Error("translator client 'Google Translate' failed to close", "error", err)
		return
	}
	slog.Debug("translator client 'Google Translate' closed")
}
