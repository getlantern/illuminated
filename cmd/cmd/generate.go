package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/getlantern/illuminated"
	"github.com/spf13/cobra"
)

var (
	projectDir  string
	source      string   // source document(s): directory or GitHub wiki URL
	targetLangs []string // target languages (ISO 639-1 codes)
	baseLang    string   // base language of source files (ISO 639-1 code)
	join        bool     // join HTML files into single document or split into individual files?
	html        bool     // generate HTML output
	pdf         bool     // generate PDF output
	// mk        bool     // generate markdown output (TODO: implement)
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate documents from source files",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:
	//
	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) { Init() },
	RunE: func(cmd *cobra.Command, args []string) error {
		// stage files from remote or outside dir to projectDir
		err := illuminated.Stage(source, projectDir)
		if err != nil {
			slog.Error("unable to stage selected source", "error", err)
			os.Exit(1)
		}
		slog.Debug("source files staged", "source", source, "projectDir", projectDir)

		// process the markdown into html
		files, err := os.ReadDir(path.Join(projectDir, illuminated.DefaultDirNameStaging))
		if err != nil {
			return fmt.Errorf("read staging directory: %w", err)
		}
		// ensure output directory exists
		err = os.MkdirAll(path.Join(projectDir, illuminated.DefaultDirNameOutput), illuminated.DefaultFilePermissions)
		if err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				slog.Debug("skipping dir (expects only files)", "name", file.Name())
				continue
			}
			sourcePath := filepath.Join(projectDir, illuminated.DefaultDirNameStaging, file.Name())
			slog.Debug("reading markdown file", "path", sourcePath)
			outName := strings.TrimSuffix(file.Name(), ".md")
			outName = fmt.Sprintf("%s.%s.%s", baseLang, outName, "html")
			outPath := path.Join(projectDir, illuminated.DefaultDirNameOutput, outName)

			err := illuminated.MarkdownToHTML(sourcePath, outPath)
			if err != nil {
				return fmt.Errorf("reading markdown file %q: %w", sourcePath, err)
			}
			slog.Debug("created HTML", "doc", sourcePath)
		}
		if join {
			// join all HTML files into one
			for _, lang := range targetLangs {
				joinedFile, err := illuminated.JoinHTML(lang, projectDir, "joined")
				if err != nil {
					return fmt.Errorf("join HTML files for language %q: %w", lang, err)
				}
				slog.Debug("joined HTML files", "file", joinedFile)
			}
		}

		// optional: output pdf
		// optional: output markdown
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// source
	generateCmd.PersistentFlags().StringVarP(
		&source, "source", "s", "",
		"source document(s) location, can be: directory, or GitHub wiki URL",
	)
	generateCmd.MarkPersistentFlagRequired("source")

	// languages
	generateCmd.PersistentFlags().StringVarP(
		&baseLang, "base", "b", "en",
		"language (ISO 639-1 code) of source files",
	)
	generateCmd.PersistentFlags().StringSliceVarP(
		&targetLangs, "languages", "l", []string{},
		"target languages (ISO 639-1 codes)",
	)
	generateCmd.MarkPersistentFlagRequired("languages")

	// output
	generateCmd.PersistentFlags().BoolVarP(&join, "join", "j", false, "join all documents into one")
	generateCmd.PersistentFlags().BoolVarP(&html, "html", "H", false, "generate HTML output")
	generateCmd.PersistentFlags().BoolVarP(&pdf, "pdf", "P", false, "generate PDF output")
	// generateCmd.PersistentFlags().BoolVarP(&mk, "markdown", "M", false, "generate markdown output") // TODO: implement
	generateCmd.MarkFlagsOneRequired("html", "pdf")
	// logging
	generateCmd.PersistentFlags().BoolVarP(&force, "force", "f",
		false,
		"overwrite existing files",
	)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
