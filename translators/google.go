package translators

import (
	"context"
	"fmt"
	"log/slog"

	g "cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

// googleTranslator implements the Translator interface using Google Translate API.
type googleTranslator struct {
	Client *g.Client
}

// NewGoogleTranslator returns a new Google Translate client.
func NewGoogleTranslator(ctx context.Context) (*googleTranslator, error) {
	// TODO: use non-local ADC credential setup for production
	client, err := g.NewClient(ctx)
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
func (g *googleTranslator) Translate(
	ctx context.Context,
	targetLang string,
	texts []string,
) ([]string, error) {
	tag := language.Make(targetLang)
	slog.Debug("making language tag for target lang", "tag", tag)
	t, err := g.Client.Translate(
		ctx,
		texts,
		tag,
		nil,
	)
	if err != nil {
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
