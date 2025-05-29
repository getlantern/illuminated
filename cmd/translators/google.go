package translators

import (
	"context"
	"fmt"
	"log/slog"

	g "cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

// TODO: define interface
type Translator interface{}

type googleTranslator struct {
	Client *g.Client
}

func NewGoogleTranslator(ctx context.Context) (*googleTranslator, error) {
	// TODO: use non-local ADC credential setup
	client, err := g.NewClient(ctx)
	if err != nil {
		return &googleTranslator{}, fmt.Errorf("create Google Translate client: %w", err)
	}
	return &googleTranslator{Client: client}, nil
}

func (g *googleTranslator) TranslateWithGoogle(ctx context.Context) error {
	g, err := NewGoogleTranslator(ctx)
	if err != nil {
		return fmt.Errorf("create Google translator: %w", err)
	}
	defer g.Client.Close()

	tag := language.Spanish
	t, err := g.Client.Translate(
		ctx,
		[]string{"hello, world"},
		tag,
		nil,
	)
	if err != nil {
		return fmt.Errorf("translate text: %w", err)
	}
	slog.Info("translated", "result", t)
	return nil
}
