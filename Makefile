# goandbackup Makefile

BINARY_NAME := goandbackup
MAIN_PKG := ./cmd/goandbackup
DIST_DIR := dist

ifeq ($(OS),Windows_NT)
	EXE_EXT := .exe
	NULL_DEVICE := nul
	DATE_CMD := powershell -NoProfile -NonInteractive -Command "Get-Date -Format 'yyyy-MM-dd_HH:mm:ss'"
	RM := del /Q
	RMDIR := rmdir /S /Q
else
	EXE_EXT :=
	NULL_DEVICE := /dev/null
	DATE_CMD := date -u +"%Y-%m-%d_%H:%M:%S"
	RM := rm -f
	RMDIR := rm -rf
endif

VERSION ?= $(shell git describe --tags --match "v[0-9]*.[0-9]*.[0-9]*" --always 2>$(NULL_DEVICE) || echo v0.0.0-dev)
COMMIT := $(shell git rev-parse --short HEAD 2>$(NULL_DEVICE) || echo unknown)
BUILD_TIME := $(shell $(DATE_CMD))

LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(BUILD_TIME)
BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

.PHONY: help
help:
	@echo goandbackup Makefile
	@echo.
	@echo "make build      Build binary into dist/"
	@echo "make test       Run tests"
	@echo "make fmt        Format Go code"
	@echo "make tidy       Tidy modules"
	@echo "make clean      Remove build artifacts"
	@echo "make run        Run CLI"
	@echo "make version    Print build metadata"

.PHONY: build
build:
	@mkdir -p $(DIST_DIR)
	go build $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)$(EXE_EXT) $(MAIN_PKG)

.PHONY: test
test:
	go test -count=1 ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: run
run:
	go run $(MAIN_PKG)

.PHONY: clean
clean:
	-@$(RMDIR) $(DIST_DIR) 2>$(NULL_DEVICE)

.PHONY: version
version:
	@echo Version: $(VERSION)
	@echo Commit:  $(COMMIT)
	@echo Built:   $(BUILD_TIME)
