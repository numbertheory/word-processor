APP_NAME := word-processor
INSTALL_DIR := $(HOME)/.local/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X word-processor/internal/editor.AppVersion=$(VERSION)"

ifeq ($(OS),Windows_NT)
	BIN := $(APP_NAME).exe
else
	BIN := $(APP_NAME)
endif

.PHONY: all build run test install clean

all: build

build:
	go build $(LDFLAGS) -o $(BIN) .

run: build
	./$(BIN)

test:
	go test -v ./...

install: build
	install -d $(INSTALL_DIR)
	install -m 755 $(BIN) $(INSTALL_DIR)/$(BIN)

clean:
	rm -f $(APP_NAME) $(APP_NAME).exe
