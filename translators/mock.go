package translators

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"golang.org/x/net/html"
)

// mockTranslator allows for unit testing mock calls to a translation service.
// Network calls are behind an "integration" build tag.
type mockTranslator struct{}

func (m *mockTranslator) SupportedLanguages(ctx context.Context, baseLang string) ([]string, error) {
	return []string{"en", "es", "ru", "fa", "ar", "zh"}, nil
}

func (m *mockTranslator) Translate(ctx context.Context, targetLang string, texts []string) ([]string, error) {
	translations := make([]string, len(texts))
	for i, text := range texts {
		doc, err := html.Parse(strings.NewReader(text))
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML: %w", err)
		}

		var substituteText func(*html.Node)
		substituteText = func(n *html.Node) {
			if n.Type == html.TextNode {
				translatedText, ok := loremIpsum[targetLang]
				if ok {
					n.Data = translatedText
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				substituteText(c)
			}
		}
		substituteText(doc)

		var buf strings.Builder
		if err := html.Render(&buf, doc); err != nil {
			return nil, fmt.Errorf("failed to render HTML: %w", err)
		}
		translations[i] = buf.String()
	}
	return translations, nil
}

func (m *mockTranslator) Close(ctx context.Context) {}

var loremIpsum = map[string]string{
	"en": "The quick brown fox jumps over the lazy dog.",
	"es": "El veloz zorro marrón salta sobre el perro perezoso.",
	"ru": "Быстрая коричневая лиса прыгает через ленивую собаку.",
	"fa": `روباه قهوه‌ای سریع از روی سگ تنبل می‌پرد.`,
	"ar": "الثعلب البني السريع يقفز فوق الكلب الكسول.",
	"zh": "快速的棕色狐狸跳过懒狗。",
}

func randWords(lang string) (string, error) {
	text, ok := loremIpsum[lang]
	if !ok {
		return "", fmt.Errorf("no lorem ipsum text available for language: %s", lang)
	}
	length := len(text)
	if length == 0 {
		return "", fmt.Errorf("lorem ipsum text for language %q is empty", lang)
	}

	start := rand.Intn(length)
	end := start + rand.Intn(length-start) + 1

	return text[start:end], nil
}
