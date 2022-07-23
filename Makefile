.POSIX:

include config.mk

all: clean build

build:
	go build ./cmd/memex/ $(GOFLAGS)

test:
	go test ./...

clean:
	rm -f memex

install: build
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	cp -f memex $(DESTDIR)$(PREFIX)/bin
	chmod 755 $(DESTDIR)$(PREFIX)/bin/memex

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/memex

.PHONY: all build test clean install uninstall
