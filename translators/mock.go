package translators

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// mockTranslator allows for unit testing mock calls to a translation service.
// Network calls are behind an "integration" build tag.
type mockTranslator struct{}

func (m *mockTranslator) SupportedLanguages(ctx context.Context, baseLang string) ([]string, error) {
	return []string{"en", "es", "ru", "fa", "zh"}, nil
}

func (m *mockTranslator) Translate(ctx context.Context, targetLang string, texts []string) ([]string, error) {
	translations := make([]string, len(texts))
	for i := range texts {
		translations[i] = loremIpsum[targetLang]
		// fake, err := randWords(targetLang)
		// if err != nil {
		// 	return nil, fmt.Errorf("generate random words for language %q: %w", targetLang, err)
		// }
		// if len(fake) == 0 {
		// 	// NOTE: this should maybe just skip silently?
		// 	return nil, fmt.Errorf("generated text for language %q is empty", targetLang)
		// }
		// translations[i] = strings.TrimSpace(fake)
	}
	return translations, nil
}

func (m *mockTranslator) Close(ctx context.Context) {}

var loremIpsum = map[string]string{
	"en": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
	"es": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore y dolor magna aliqua.",
	"ru": "Лорем ипсум долор сит амет, консектетур адиписцинг элит, сед до еиусмод темпор инцидидунт ут лаборе ет долоре магна аликуа.",
	"fa": "لورم ایپسوم متن ساختگی با تولید سادگی نامفهوم از صنعت چاپ و با استفاده از طراحان گرافیک استu",
	"ar": "لوريم ايبسوم دولار سيت أميت , كونسيكتيتور أديبيسكينغ أليت , سيد دو أيوسمود تيمبور إنسيديدونت أوت لابوري إت دولار ماجنا أليكوا.",
	"zh": "假文本文是印刷和排版行业的虚拟文本。",
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

	rand.Seed(time.Now().UnixNano())
	start := rand.Intn(length)
	end := start + rand.Intn(length-start)

	return text[start:end], nil
}
