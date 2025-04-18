package illuminated

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

var (
	ErrSourceValidation = fmt.Errorf("source must be a valid file, directory, or GitHub wiki")
)

// Stage validates source and stages files in a staging directory,
// optionally deleting the files on completion if cleanup is true.
func Stage(source string, projectDir string) error {
	parsedURL, err := url.Parse(source)
	if err == nil && parsedURL.Scheme != "" && parsedURL.Host != "" {
		slog.Debug("fetching remote wiki", "URL", parsedURL)
		// TODO fetch remote wiki
		return fmt.Errorf("remote GitHub wiki URL fetching not implemented")
	}

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
