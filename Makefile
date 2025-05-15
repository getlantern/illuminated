.PHONY: all build run

all: build

build:
	cd cmd && go build -o ../illuminated

testlocal:
	./illuminated cleanup --force --verbose
	./illuminated init --verbose
	./illuminated update --source example --verbose
	./illuminated generate --html --verbose

testremote:
	./illuminated cleanup --force --verbose
	./illuminated init --verbose
	./illuminated update --source https://github.com/getlantern/guide.wiki.git --verbose
	./illuminated generate --join --html --verbose
