.PHONY: all build run

all: build

build:
	cd cmd && go build -o ../illuminated

testlocal:
	./illuminated cleanup --force
	./illuminated init
	./illuminated update --source example
	./illuminated generate --join --pdf

testremote:
	./illuminated cleanup --force
	./illuminated init
	./illuminated update --source https://github.com/getlantern/guide.wiki.git
	./illuminated generate --join --pdf
