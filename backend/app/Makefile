BINARY := app

.PHONY: all generate build clean run

all: clear generate build run

clear:
	@clear

generate:
	@echo "Generating templ files..."
	go tool templ generate ./...

build: generate
	@echo "Building binary..."
	CGO_ENABLED=0 go build -o $(BINARY) .

run: build
	@echo "Starting $(BINARY)..."
	./$(BINARY)

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY)
