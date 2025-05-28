package translators

import (
	"context"
	"fmt"
	"log/slog"

	g "cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

func TranslateWithGoogle() error {
	ctx := context.Background()
	client, err := g.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("create Google Translate client: %w", err)
	}
	defer client.Close()

	lang := g.Language{
		Name: "Spanish",
		Tag:  language.Spanish,
	}
	t, err := client.Translate(
		ctx,
		[]string{"hello, world"},
		lang.Tag,
		nil,
	)
	if err != nil {
		return fmt.Errorf("translate text: %w", err)
	}
	slog.Info("translated", "result", t)
	return nil
}
