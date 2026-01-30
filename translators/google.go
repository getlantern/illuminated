package translators

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	gtx "cloud.google.com/go/translate"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

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
		return &googleTranslator{}, fmt.Errorf("create Google Translate client: %w", err)
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
// If translation fails with a 400 error and content is large, it will retry with chunked translation.
func (g *googleTranslator) Translate(
	ctx context.Context,
	targetLang string,
	texts []string,
) ([]string, error) {
	tag := language.Make(targetLang)
	// slog.Debug("making language tag for target lang", "tag", tag) // TODO: enable
	slog.Debug("making language tag for target lang", "tag", tag, "texts", texts) // FIXME: redundant
	t, err := g.Client.Translate(
		ctx,
		texts,
		tag,
		&gtx.Options{Format: gtx.HTML},
	)
	if err != nil {
		slog.Error("translation failed",
			"targetLang", targetLang,
			"textLength", len(texts[0]),
			"numTexts", len(texts),
			"error", err,
		)

		// If it's a 400 error and content is reasonably large, try chunking
		if len(texts) == 1 && len(texts[0]) > 2000 {
			slog.Info("retrying with chunked translation", "originalSize", len(texts[0]))
			return g.translateChunked(ctx, targetLang, texts[0])
		}

		return nil, fmt.Errorf("translate text: %w", err)
	}
	for i, translation := range t {
		slog.Debug("translated text",
			"index", i,
			"source", translation.Source,
			"target", translation.Model,
			"text", translation.Text,
		)
	}
	var translatedTexts []string
	for _, translation := range t {
		translatedTexts = append(translatedTexts, translation.Text)
	}
	return translatedTexts, nil
}

// translateChunked splits HTML content by h2 sections and translates each chunk separately.
func (g *googleTranslator) translateChunked(ctx context.Context, targetLang string, html string) ([]string, error) {
	tag := language.Make(targetLang)

	// Split by h2 tags to create manageable chunks
	chunks := splitByH2(html)
	slog.Debug("split into chunks", "count", len(chunks))

	var translatedChunks []string
	for i, chunk := range chunks {
		slog.Debug("translating chunk", "index", i, "size", len(chunk))
		t, err := g.Client.Translate(
			ctx,
			[]string{chunk},
			tag,
			&gtx.Options{Format: gtx.HTML},
		)
		if err != nil {
			return nil, fmt.Errorf("translate chunk %d: %w", i, err)
		}
		translatedChunks = append(translatedChunks, t[0].Text)
	}

	// Join all translated chunks
	result := ""
	for _, chunk := range translatedChunks {
		result += chunk
	}

	return []string{result}, nil
}

// splitByH2 splits HTML content into chunks at h2 boundaries.
func splitByH2(html string) []string {
	// Simple split by <h2> tags
	parts := regexp.MustCompile(`(?i)(<h2>)`).Split(html, -1)
	if len(parts) <= 1 {
		// No h2 tags, return as single chunk
		return []string{html}
	}

	var chunks []string
	// First part (before first h2)
	if len(parts[0]) > 0 {
		chunks = append(chunks, parts[0])
	}

	// Remaining parts (need to prepend <h2> back)
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
