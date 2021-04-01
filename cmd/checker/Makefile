.PHONY: build pub aaa

BIN = "checker"
SERVICE_NAME = "security-checker"
BUILD_DIR = build
PUB_DIR = pub

INSTALL_DIR = "/usr/local/security-checker"

VERSION := $(shell git describe --always --tags)
DATE := $(shell date -u +'%Y-%m-%d %H:%M:%S')
GOVERSION := $(shell go version)
COMMIT := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMITER := $(shell git log -1 --pretty=format:'%an')
UPLOADER:= $(shell hostname)/${USER}/${COMMITER}

define GIT_INFO
//nolint
package git
const (
	BuildAt string="$(DATE)"
	Version string="$(VERSION)"
	Golang string="$(GOVERSION)"
	Commit string="$(COMMIT)"
	Branch string="$(BRANCH)"
	Uploader string="$(UPLOADER)"
);
endef
export GIT_INFO


define build
	@echo "===== building $(BIN) for $(1)-$(2) ===="
	@mkdir -p "$(BUILD_DIR)/$(1)-$(2)" "$(BUILD_DIR)/pub"
	@GOOS=$(1) GOARCH=$(2) go build -o $(BUILD_DIR)/$(1)-$(2)/$(BIN) -ldflags "-w -s -X main.Version=$(VERSION)" main.go
	@tar -czf $(BUILD_DIR)/pub/$(SERVICE_NAME)-$(1)-$(2)-$(VERSION).tar.gz autostart -C "$(BUILD_DIR)/$(1)-$(2)" $(BIN)
	@cp -f install.sh $(BUILD_DIR)/pub/install.sh
endef

build:
	@rm -f "$(BUILD_DIR)/pub"/*
	$(call build,linux,amd64)
	$(call build,linux,386)
	@tree -Csh -L 3 $(BUILD_DIR)


local:
	$(call build,linux,amd64)
	@cp build/linux-amd64/$(BIN) /usr/local/security-checker/