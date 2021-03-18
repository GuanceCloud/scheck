.PHONY: default

default: local

BIN = "sec-checker"
BUILD_DIR = build

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
	@echo $(pwd)
	@echo "===== bilding $(BIN) ===="
	@mkdir -p $(BUILD_DIR)
	@mkdir -p git
	@echo "$$GIT_INFO" > git/git.go
	@GO111MODULE=off CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BIN) -ldflags "-w -s" cmd/sec-checker/main.go
	@tree -Csh -L 3 $(BUILD_DIR)
endef


local:
	$(call build)

	