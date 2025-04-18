.PHONY: all build run

all: build run

build:
	cd cmd && go build -o ../illuminated

run: 
	echo "cleanup..."
	./illuminated cleanup
	echo "init..."
	./illuminated init
	echo "prepare..."
	./illuminated prepare -s example
