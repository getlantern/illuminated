package translators

import (
	"context"
	"fmt"
)

const (
	TranslatorGoogle = "google"
	TranslatorMock   = "mock"
)

var ValidTranslators = []string{TranslatorGoogle}

// Translator is an interface for a generic translator.
type Translator interface {
	SupportedLanguages(ctx context.Context, baseLang string) ([]string, error)
	Translate(ctx context.Context, targetLang string, texts []string) ([]string, error)
	Close(ctx context.Context)
}

// NewTranslator returns a pointer to a new, specified translatorType
// to satisfy the Translator interface.
func NewTranslator(ctx context.Context, translatorType string) (Translator, error) {
	switch translatorType {
	case TranslatorGoogle:
		return NewGoogleTranslator(ctx)
	case TranslatorMock:
		return &mockTranslator{}, nil
	default:
		return nil, fmt.Errorf(
			"unknown translator type; given: %s, expected: %v",
			translatorType,
			ValidTranslators,
		)
	}
}
