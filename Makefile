APP_NAME := word-processor

ifeq ($(OS),Windows_NT)
	BIN := $(APP_NAME).exe
else
	BIN := $(APP_NAME)
endif

.PHONY: all build run test clean

all: build

build:
	go build -o $(BIN) .

run: build
	./$(BIN)

test:
	go test -v ./...

clean:
	rm -f $(APP_NAME) $(APP_NAME).exe
