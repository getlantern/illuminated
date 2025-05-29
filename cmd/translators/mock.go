package translators

import (
	"context"
	"fmt"
)

// mockTranslator allows for unit testing mock calls to a translation service.
// Network calls are behind an "integration" build tag.
type mockTranslator struct{}

func (m *mockTranslator) SupportedLanguages(ctx context.Context, baseLang string) ([]string, error) {
	return []string{"en", "es"}, nil
}

func (m *mockTranslator) Translate(ctx context.Context, targetLang string, texts []string) ([]string, error) {
	translations := make([]string, len(texts))
	for i, text := range texts {
		translations[i] = fmt.Sprintf("this would be %q, but in %s", text, targetLang)
	}
	return translations, nil
}

func (m *mockTranslator) Close(ctx context.Context) {}
