package illuminated

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

var (
	ErrSourceValidation = fmt.Errorf("source must be a valid file, directory, or GitHub wiki")
)

// Stage prepares the source files for processing by copying them to illuminated.DefaultDirNameStaging.
// Accepted source includes: local file, directories, or remote GitHub wiki URLs.
func Stage(source string, projectDir string) error {
	parsedURL, err := url.Parse(source)
	if err == nil && parsedURL.Scheme != "" && parsedURL.Host != "" {
		slog.Debug("staging remote wiki", "URL", parsedURL)
		err = cloneRepo(source, path.Join(projectDir, DefaultDirNameStaging))
		if err != nil {
			return fmt.Errorf("clone repo: %v", err)
		}
	} else {
		slog.Debug("staging local source", "source", source)
		err = os.MkdirAll(path.Join(projectDir, DefaultDirNameStaging), os.ModePerm)
		if err != nil {
			return fmt.Errorf("create staging: %v", err)
		}
		info, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrSourceValidation, err)
		}
		if info.IsDir() {
			entries, err := os.ReadDir(source)
			if err != nil {
				return fmt.Errorf("read dir: %v", err)
			}
			for _, entry := range entries {
				if entry.IsDir() {
					slog.Debug("ignoring directory", "name", entry.Name())
					continue
				}
				err = copy(
					filepath.Join(source, entry.Name()),
					filepath.Join(projectDir, DefaultDirNameStaging, entry.Name()),
				)
				if err != nil {
					return fmt.Errorf("stage file %q from dir: %v", entry.Name(), err)
				}
			}
		} else {
			err = copy(
				source,
				filepath.Join(projectDir, DefaultDirNameStaging, filepath.Base(source)),
			)
			if err != nil {
				return fmt.Errorf("stage single file: %v", err)
			}
		}
	}
	if err != nil {
		return ErrSourceValidation
	}
	return nil
}

// copy a single file from src to dst
func copy(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create file: %v", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("copy file: %v", err)
	}
	return nil
}

// cloneRepo shallow clones a Git repository from the given URL to the specified path.
func cloneRepo(url, path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		slog.Warn("repo already exists, replacing", "path", path)
		err = os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("failed to remove existing directory: %v", err)
		}
	}

	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}
	return nil
}
