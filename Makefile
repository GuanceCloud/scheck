.PHONY: default

default: local

# 正式环境
RELEASE_DOWNLOAD_ADDR = zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/sec-checker

# 测试环境
TEST_DOWNLOAD_ADDR = zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/sec-checker

PUB_DIR = dist
BUILD_DIR = dist

LOCAL_ARCHS = "local"
DEFAULT_ARCHS = "all"

BIN = sec-checker
NAME = sec-checker
ENTRY = cmd/sec-checker/main.go


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
	@echo "===== $(BIN) $(1) ===="
	@rm -rf $(PUB_DIR)/$(1)/*
	@mkdir -p $(BUILD_DIR) $(PUB_DIR)/$(1)
	@mkdir -p git
	@echo "$$GIT_INFO" > git/git.go
	@GO111MODULE=off CGO_ENABLED=0 go run cmd/make/make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir $(BUILD_DIR) \
		 -env $(1) -pub-dir $(PUB_DIR) -archs $(2) -download-addr $(3)
	@tree -Csh -L 3 $(BUILD_DIR)
endef

local:
	$(call build,local, $(LOCAL_ARCHS), $(LOCAL_DOWNLOAD_ADDR))

testing:
	$(call build,test, $(DEFAULT_ARCHS), $(TEST_DOWNLOAD_ADDR))

release:
	$(call build,release, $(DEFAULT_ARCHS), $(RELEASE_DOWNLOAD_ADDR))