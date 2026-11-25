SHELL := /bin/sh
.DEFAULT_GOAL := help

WAILS_VERSION := v2.10.2
GO_BIN := $(shell go env GOBIN)
ifeq ($(strip $(GO_BIN)),)
GO_BIN := $(shell go env GOPATH)/bin
endif
WAILS := $(GO_BIN)/wails
UNAME_S := $(shell uname -s)
XDG_CONFIG_HOME ?= $(HOME)/.config
MEDIA_TOOLS_DIR := $(XDG_CONFIG_HOME)/Yumdown/tools
ifeq ($(UNAME_S),Darwin)
MEDIA_TOOLS_DIR := $(HOME)/Library/Application Support/Yumdown/tools
endif

# Fedora 40+ ships WebKitGTK 4.1; older Linux builds use WebKitGTK 4.0.
WAILS_TAGS := $(shell if command -v pkg-config >/dev/null 2>&1 && pkg-config --exists webkit2gtk-4.0; then :; elif command -v pkg-config >/dev/null 2>&1 && pkg-config --exists webkit2gtk-4.1; then printf '%s' '-tags webkit2_41'; fi)

.PHONY: help install-deps install-system-deps install-go-deps install-frontend-deps install-wails-cli install-media-tools check-wails dev build test test-go test-frontend clean

help:
	@printf '%s\n' \
	  'make install-deps       Install native, Go, frontend, Wails, and media-tool dependencies' \
	  'make install-media-tools Link installed yt-dlp and ffmpeg into Yumdown config' \
	  'make dev                Run the Wails development app (auto-selects WebKitGTK tag)' \
	  'make build              Build the Wails desktop app' \
	  'make test               Run Go and frontend tests' \
	  'make clean              Remove generated frontend and Wails build output'

install-deps:
	$(MAKE) install-system-deps
	$(MAKE) install-go-deps
	$(MAKE) install-frontend-deps
	$(MAKE) install-wails-cli
	$(MAKE) install-media-tools

install-system-deps:
ifeq ($(UNAME_S),Linux)
	@. /etc/os-release; case "$$ID" in \
		fedora) sudo dnf install -y gcc-c++ pkgconf-pkg-config gtk3-devel webkit2gtk4.1-devel libsoup3-devel ffmpeg-free yt-dlp ;; \
		ubuntu|debian) sudo apt-get update && sudo apt-get install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.0-dev ffmpeg yt-dlp ;; \
		*) printf 'Unsupported Linux distribution: %s\n' "$$ID"; exit 1 ;; \
	esac
else ifeq ($(UNAME_S),Darwin)
	@printf '%s\n' 'macOS native prerequisites are Xcode Command Line Tools; install them with: xcode-select --install'
else
	@printf '%s\n' 'Install Wails native prerequisites for your OS using the Wails documentation.'
endif

install-go-deps:
	go mod download

install-frontend-deps:
	cd frontend && pnpm install --frozen-lockfile

install-wails-cli:
	go install github.com/wailsapp/wails/v2/cmd/wails@$(WAILS_VERSION)

install-media-tools:
	@command -v yt-dlp >/dev/null 2>&1 || { printf '%s\n' 'yt-dlp is not installed; install it with your OS package manager first.'; exit 1; }
	@command -v ffmpeg >/dev/null 2>&1 || { printf '%s\n' 'ffmpeg is not installed; install it with your OS package manager first.'; exit 1; }
	@mkdir -p "$(MEDIA_TOOLS_DIR)"
	@ln -sfn "$$(command -v yt-dlp)" "$(MEDIA_TOOLS_DIR)/yt-dlp"
	@ln -sfn "$$(command -v ffmpeg)" "$(MEDIA_TOOLS_DIR)/ffmpeg"
	@if command -v ffprobe >/dev/null 2>&1; then ln -sfn "$$(command -v ffprobe)" "$(MEDIA_TOOLS_DIR)/ffprobe"; fi
	@printf 'Media tools linked into %s\n' "$(MEDIA_TOOLS_DIR)"

check-wails:
	@test -x "$(WAILS)" || { printf 'Wails CLI not found at %s; run make install-wails-cli first.\n' "$(WAILS)"; exit 1; }

dev: check-wails
	$(WAILS) dev $(WAILS_TAGS)

build: check-wails
	$(WAILS) build $(WAILS_TAGS)

test: test-go test-frontend

test-go:
	go test ./...

test-frontend:
	cd frontend && pnpm test

clean:
	rm -rf build/bin frontend/dist/assets frontend/dist/index.html
