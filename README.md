# illuminated

[![Go - Test & Build](https://github.com/getlantern/illuminated/actions/workflows/go.yml/badge.svg)](https://github.com/getlantern/illuminated/actions/workflows/go.yml)

internationalization tool for GitHub wikis

## purpose
Converts a GitHub wiki into an HTML or PDF, optionally translated into multiple languages, to support rapid iteration of GitHub Wiki content while maintaining broad internationalization support and document generation for distribution.

## dependencies
[pandoc](https://pandoc.org/) and the xelatex PDF engine

## usage

### development
To delete all example files and start over with newly built binary, run:
```sh
$ ./test.sh {local|remote} {mock|google} [comma-separated-languages]
```

### production
Build the binary.
```sh
$ go build -o illuminated ./cmd
```

Generate a single, joined HTML and PDF for 5 languages using Google translate.
```sh
$ ./illuminated generate --verbose \
  --source https://github.com/getlantern/guide.wiki.git \
  --base "en" \
  --languages "en,zh,ru,fa,ar" \
  --translator "google" \
  --overrides "../overrides.yml" \
  --title "User Guide" \
  --html \
  --pdf \
  --join \
  --force
```

Use the help command for details.
```sh
$ ./illuminated --help
```
### overrides
If a specific phrase is needed for a particular language, define that in an `overrides.yml` file in the directory where the command is run (or specify a different path with the `--overrides` flag).

Example `overrides.yml`:
```yaml
- title: Lantern
  language: zh
  original: 灯笼
  replacement: 蓝灯
- title: Block
  language: en
  original: blacklist
  replacement: block list
- title: Allow
  language: en
  original: whitelist
  replacement: allow list
```

