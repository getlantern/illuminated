package illuminated

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

// WikiIgnore defines file names which should be ignored when staging a cloned remote Wiki.
var WikiIgnore = []string{
	".git",
	"Home",
	"_Sidebar",
	"_Footer",
}

// Stage fetches new source files for processing,
// copying them to illuminated.DefaultDirNameStaging.
// Accepted sources include:
//   - local directory path
//   - GitHub wiki URL
func Stage(source string, projectDir string) error {
	parsedURL, err := url.Parse(source)
	if err == nil && parsedURL.Scheme != "" && parsedURL.Host != "" {
		slog.Debug("staging remote wiki", "URL", parsedURL)
		err = cloneRepo(source, path.Join(projectDir, DefaultDirNameStaging))
		if err != nil {
			return fmt.Errorf("clone repo: %w", err)
		}
		// remove ignored files
		dir, err := os.ReadDir(path.Join(projectDir, DefaultDirNameStaging))
		if err != nil {
			return fmt.Errorf("read staging dir: %w", err)
		}
		for _, entry := range dir {
			if entry.IsDir() {
				continue
			}
			for _, ignore := range WikiIgnore {
				if strings.Contains(entry.Name(), ignore) {
					slog.Debug("removing ignored file", "name", entry.Name())
					err = os.Remove(path.Join(projectDir, DefaultDirNameStaging, entry.Name()))
					if err != nil {
						return fmt.Errorf("remove ignored file: %w", err)
					}
				}
			}
		}
	} else {
		slog.Debug("staging local source", "source", source)
		err = os.MkdirAll(path.Join(projectDir, DefaultDirNameStaging), 0o750)
		if err != nil {
			return fmt.Errorf("create staging: %w", err)
		}
		info, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf("invalid source: %w", err)
		}
		if info.IsDir() {
			entries, err := os.ReadDir(source)
			if err != nil {
				return fmt.Errorf("read dir: %w", err)
			}
			for _, entry := range entries {
				if entry.IsDir() {
					slog.Debug("ignoring directory", "name", entry.Name())
					continue
				}
				slog.Debug("copying file", "name", entry.Name())
				err = copy(
					filepath.Join(source, entry.Name()),
					filepath.Join(projectDir, DefaultDirNameStaging, entry.Name()),
				)
				if err != nil {
					return fmt.Errorf("stage file %q from dir: %w", entry.Name(), err)
				}
			}
		} else {
			return fmt.Errorf("source is not a directory: %v", source)
		}
	}
	slog.Info("staging complete", "dir", path.Join(projectDir, DefaultDirNameStaging))
	return nil
}

// copy a single file from src to dst
func copy(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	return nil
}

// cloneRepo shallow clones a Git repository from the given URL to the specified path.
func cloneRepo(url, path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		slog.Warn("repo already exists, replacing", "path", path)
		err = os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	return nil
}
