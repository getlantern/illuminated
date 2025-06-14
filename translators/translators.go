package translators

import (
	"context"
	"fmt"
)

const (
	GoogleTranslate = "google"
	MockTranslation = "mock"
)

var ValidTranslators = []string{
	MockTranslation,
	GoogleTranslate,
}

var ErrTranslatorUnsupported = fmt.Errorf("translator not supported")

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
	case GoogleTranslate:
		return NewGoogleTranslator(ctx)
	case MockTranslation:
		return &mockTranslator{}, nil
	default:
		return nil, fmt.Errorf(
			"%w: (given: %s, expected: %v)",
			ErrTranslatorUnsupported,
			translatorType,
			ValidTranslators,
		)
	}
}
