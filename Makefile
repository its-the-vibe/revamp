.PHONY: all build install run test lint clean

BINARY=revamp

all: build

build:
	go build -o $(BINARY) .

install:
	go install .

run:
	./$(BINARY)

test:
	go test ./...

lint:
	golangci-lint run || echo "Install golangci-lint for linting"

clean:
	rm -f $(BINARY)
