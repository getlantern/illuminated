# illuminated

[![Go - Test & Build](https://github.com/getlantern/illuminated/actions/workflows/go.yml/badge.svg)](https://github.com/getlantern/illuminated/actions/workflows/go.yml)

internationalization tool for GitHub wikis

## function
Converts a GitHub wiki into an HTML or PDF document, optionally translated into multiple languages.

## purpose

Support rapid iteration of GitHub Wiki while maintaining broad internationalization support and document generation for distirbution. 

## usage

### development
To delete all example files and start over with newly built binary, run:
```sh
$ ./test.sh {local|remote} {mock|google}
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
  --html \
  --pdf \
  --join
```

Use the help command for details.
```sh
$ ./illuminated --help
```

---

## Future Work
- [ ] support style for pagebreaks
- [ ] handle footer
- [ ] don't mutate data in place
- [ ] fix bug with "skipping file with no body"
- [ ] update fonts with Noto *
- [ ] support overrides

## fonts
- DejaVu works for ru, fa, not zh

