#!/usr/bin/env bash

usage() {
    echo "usage: $0 {local|remote} {mock|google}"
}


if [[ $# -lt 2 ]]; then
  usage
  exit 1
fi

if [[ -n "$DEBUG" ]]; then
  echo "DEBUG=$DEBUG, running with --verbose"
  VERBOSE="--verbose"
fi

build() {
  echo "building ..."
  cd cmd || exit 1
  go build -o ../illuminated || exit 1
  cd ..
}

cleanup() {
  echo "cleanup ..."
  ./illuminated cleanup --force "$VERBOSE"
}

build
cleanup

case "$1" in
  local)
    SOURCE="example"
    echo "generating local ($SOURCE) ..."
    ;;
  remote)
    SOURCE="https://github.com/getlantern/guide.wiki.git"
    echo "generating remote ($SOURCE) ..."
    ;;
  *)
    usage
    exit 1
esac

case "$2" in
  mock)
    echo "translating (mock) ..."
    TRANSLATOR="mock"
    ;;
  google)
    echo "translating (google) ..."
    TRANSLATOR="google"
    ;;
  *)
    usage
    exit 1
    ;;
esac

set -x
./illuminated generate "$VERBOSE" \
  --source "$SOURCE" \
  --base "en" \
  --languages "zh" \
  --translator "$TRANSLATOR" \
  --html \
  --pdf \
  --join \
# --languages "en,zh,ru,fa,ar" \
