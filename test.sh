#!/usr/bin/env bash

build() {
  echo "building..."
  cd cmd || exit 1
  go build -o ../illuminated || exit 1
  cd ..
}

build

# case switch with "test, remote, local, and translate"
case "$1" in
  test)
    echo "testing..."
    go test ./... -v || exit 1
    ;;
  local)
    echo "generating local..."
    ./illuminated cleanup --force --verbose
    ./illuminated init --verbose
    ./illuminated update --source example --verbose
    ./illuminated generate --pdf --html --verbose
    # Add your local command logic here
    ;;
  remote)
    echo "generating remote..."
    ./illuminated cleanup --force --verbose
    ./illuminated init --verbose
    ./illuminated update --source https://github.com/getlantern/guide.wiki.git --verbose
    ./illuminated generate --pdf --html --verbose
    # Add your remote command logic here
    ;;
  translate)
    echo "translating..."
    ./illuminated cleanup --force --verbose
    # ./illuminated init --base en --target en,fa,ru,zh --verbose
    ./illuminated init --base en --target en --verbose
    ./illuminated update --source example --verbose
    ./illuminated translate --translator mock --verbose
    ./illuminated generate --html --pdf --join --verbose
    # Add your translation logic here
    ;;
  *)
    echo "usage: $0 {test|remote|local|translate}"
    exit 1
    ;;
esac
