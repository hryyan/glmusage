.PHONY: build run install clean

BINARY = glmusage

build:
	go build -o $(BINARY) ./cmd/glmusage/

run: build
	./$(BINARY)

install:
	go install ./cmd/glmusage/

clean:
	rm -f $(BINARY)
