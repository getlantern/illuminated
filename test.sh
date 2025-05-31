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

restart() {
  echo "cleanup ..."
  ./illuminated cleanup --force "$VERBOSE"
  echo "initializing ..."
  ./illuminated init --base en --target en,fa,ru,ar,zh "$VERBOSE"
}

build
restart

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
    echo "translating (google) ..."
    TRANSLATOR="mock"
    ;;
  google)
    echo "translating (mock) ..."
    TRANSLATOR="google"
    ;;
  *)
    usage
    exit 1
    ;;
esac

./illuminated update --source "$SOURCE" "$VERBOSE"
./illuminated translate --translator "$TRANSLATOR" "$VERBOSE"
./illuminated generate --pdf "$VERBOSE"
