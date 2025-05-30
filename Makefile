SRC := $(shell find cmd -type f -name '*.go')
BINARY := illuminated

# Default target
all: $(BINARY)

# Build the binary if source files change
$(BINARY): $(SRC)
	cd cmd && go build -o ../$(BINARY)
	
# Rebuild only
rebuild: $(BINARY)
	./$(BINARY) --help

# Test local files 
local: $(BINARY)
	./$(BINARY) cleanup --force --verbose
	./$(BINARY) init --verbose
	./$(BINARY) update --source example --verbose
	./$(BINARY) generate --pdf --join --verbose

# Test remote files from a wiki URL
remote: $(BINARY)
	./$(BINARY) cleanup --force --verbose
	./$(BINARY) init --verbose
	./$(BINARY) update --source https://github.com/getlantern/guide.wiki.git --verbose
	./$(BINARY) generate --pdf --join --verbose
	
# Clean up the binary
cleanup:
	./$(BINARY) cleanup --force --verbose
	rm -f $(BINARY)

# translate: $(BINARY)
# 	./$(BINARY) cleanup --force --verbose
# 	./$(BINARY) init --verbose
# 	./$(BINARY) update --source example --verbose
# 	./$(BINARY) translate --verbose

translateq: $(BINARY)
	./$(BINARY) cleanup --force 
	./$(BINARY) init --base en --target en,fa,ru 
	./$(BINARY) update --source example 
	./$(BINARY) translate --translator mock 
