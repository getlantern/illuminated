.PHONY: all build run

all: build run

build:
	cd cmd && go build -o ../illuminated

run: 
	echo "cleanup..."
	./illuminated cleanup --force
	echo "init..."
	./illuminated init
	echo "prepare..."
	./illuminated prepare -s https://github.com/getlantern/guide.wiki.git
