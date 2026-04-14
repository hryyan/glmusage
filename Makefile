.PHONY: build run install clean

build:
	go build -o glmusage ./cmd/glmusage/

run: build
	./glmusage

install:
	go install ./cmd/glmusage/

clean:
	rm -f glmusage
