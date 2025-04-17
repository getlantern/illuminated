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
	DefaultConfig         = config{
		Source:  "en",
		Targets: []string{"en", "zh"},
	}
)

type config struct {
	Source  string   `json:"source"`
	Targets []string `json:"target"`
}

func (c *config) write(dir string) error {
	// check for existing config file
	if _, err := os.Stat(path.Join(dir, DefaultConfigFilename)); err == nil {
		slog.Warn("existing config exists and will be overwritten", "file", DefaultConfigFilename)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check config file %v: %v", DefaultConfigFilename, err)
	}

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create directory %v: %v", dir, err)
	}
	configPath := path.Join(dir, DefaultConfigFilename)
	slog.Debug("creating config file", "file", DefaultConfigFilename)
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("create config file %v: %v", DefaultConfigFilename, err)
	}
	defer f.Close()

	yamlData, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %v", err)
	}

	err = os.WriteFile(configPath, yamlData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("write config file %v: %v", DefaultConfigFilename, err)
	}

	return nil
}

func (c *config) read(filepath string) error {
	f, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("read config file %v: %v", filepath, err)
	}

	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return fmt.Errorf("unmarshal config file %v: %v", filepath, err)
	}

	if c.Source == "" {
		slog.Warn("source language not set, using default", "lang", DefaultConfig.Source)
		c.Source = DefaultConfig.Source
	}
	if len(c.Targets) == 0 {
		slog.Warn("target languages not set, using default", "langs", DefaultConfig.Targets)
		c.Targets = DefaultConfig.Targets
	}
	return nil
}
