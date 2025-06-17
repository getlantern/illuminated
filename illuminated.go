package illuminated

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	DefaultFilePermissions = os.FileMode(0o750)
	DefaultDirNameStaging  = "staging"
	DefaultDirNameOutput   = "output"
)

// override defines a word or phrase that should be overridden if/when it exists in a translation.
type override struct {
	Title       string `yaml:"title,omitempty"`
	Language    string `yaml:"language,omitempty"`
	Original    string `yaml:"original,omitempty"`
	Replacement string `yaml:"replacement,omitempty"`
}

// writeOverrides writes a slice of overrides to a YAML file at path.
func writeOverrides(path string, overrides []override) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create override file")
	}
	defer f.Close()
	encoder := yaml.NewEncoder(f)
	defer encoder.Close()

	err = encoder.Encode(overrides)
	if err != nil {
		return fmt.Errorf("encode overrides: %w", err)
	}
	slog.Debug(
		"overrides written to file",
		"count", len(overrides),
		"path", path,
	)
	return nil
}

// readOverrides reads a slice of overrides from a YAML file at path.
func readOverrides(path string) ([]override, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open override file: %w", err)
	}
	defer f.Close()
	var overrides []override
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&overrides)
	if err != nil {
		return nil, fmt.Errorf("decode overrides: %w", err)
	}
	slog.Debug(
		"overrides read from file",
		"count", len(overrides),
		"path", path,
	)
	return overrides, nil
}
