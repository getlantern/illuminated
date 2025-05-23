package illuminated

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

var (
	DefaultConfigFilename = ".illuminatedrc"
	DefaultConfig         = Config{
		Base:    "en",
		Targets: []string{"en", "fa"},
	} // CLI has its own defaults which may supercede these
	DefaultDirNameStaging      = "staging"      // copies of source and intermediate files
	DefaultDirNameTranslations = "translations" // translation files for internationalization
	DefaultDirNameTemplates    = "templates"    // template to recreate localized copies
	DefaultDirNameOutput       = "output"       // final output (typically PDF)
	DefaultFilePermissions     = os.FileMode(0o750)
	ErrNoClobber               = fmt.Errorf(
		"config file already exists",
	)
)

// Config defines the base language from which all translations will be derived,
// and all languages that will be translated (assumes ISO 639-1 codes).
type Config struct {
	Base    string   `json:"base"`   // original source language
	Targets []string `json:"target"` // translated languages
}

// Write creates a config file in the specified directory,
// overwriting any existing config if force=true.
func (c *Config) Write(dir string, force bool) error {
	filepath := path.Join(dir, DefaultConfigFilename)
	if _, err := os.Stat(filepath); err == nil {
		if !force {
			return fmt.Errorf("%s: %w", DefaultConfigFilename, ErrNoClobber)
		}
		slog.Info("existing config exists and will be overwritten", "filepath", filepath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check config file %v: %w", DefaultConfigFilename, err)
	}

	err := os.MkdirAll(dir, DefaultFilePermissions)
	if err != nil {
		return fmt.Errorf("create directory %v: %w", dir, err)
	}
	configPath := path.Join(dir, DefaultConfigFilename)
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("create config file %v: %w", DefaultConfigFilename, err)
	}
	defer f.Close()

	yamlData, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = os.WriteFile(configPath, yamlData, DefaultFilePermissions)
	if err != nil {
		return fmt.Errorf("write config file %v: %w", DefaultConfigFilename, err)
	}
	slog.Info("project directory created with config", "dir", dir, "config", configPath)
	return nil
}

// Read reads a config file from the specified path.
func (c *Config) Read(filepath string) error {
	f, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("read config file %v: %w", filepath, err)
	}

	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return fmt.Errorf("unmarshal config file %v: %w", filepath, err)
	}

	if c.Base == "" {
		slog.Warn("source language not set, using default", "lang", DefaultConfig.Base)
		c.Base = DefaultConfig.Base
	}
	if len(c.Targets) == 0 {
		slog.Warn("target languages not set, using default", "langs", DefaultConfig.Targets)
		c.Targets = DefaultConfig.Targets
	}
	return nil
}
