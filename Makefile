.PHONY: build pub man

BIN = "scheck"
BUILD_DIR = build
PUB_DIR = pub
ENTRY = cmd/checker/main.go

# 正式环境
RELEASE_DOWNLOAD_ADDR = zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/security-checker

# 测试环境
TEST_DOWNLOAD_ADDR = zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker


LOCAL_ARCHS = "local"
LOCAL_DOWNLOAD_ADDR = $(LOCAL_OSS_BUCKET)"."$(LOCAL_OSS_HOST)"/"$(hostname)"/"security-checker
DEFAULT_ARCHS = "all"

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
	@echo "===== $(BIN) $(1) ===="
	@echo "==git version=== $(VERSION) ===="
	@echo "=== $(TEST_DOWNLOAD_ADDR) ===="
	@rm -rf $(PUB_DIR)/$(1)/*
	@mkdir -p $(BUILD_DIR) $(PUB_DIR)/$(1)
	@mkdir -p git
	@echo "$$GIT_INFO" > git/git.go
	@GO111MODULE=off CGO_ENABLED=0 go run cmd/make/make.go -main $(ENTRY) -binary $(BIN) -build-dir $(BUILD_DIR) \
		 -env $(1) -pub-dir $(PUB_DIR) -archs $(2) -download-addr $(3)

endef

define pub
	@echo "publish $(1) ..."
	@GO111MODULE=off go run cmd/make/make.go -pub -env $(1) -pub-dir $(PUB_DIR) -binary $(BIN) -download-addr $(2) \
		-build-dir $(BUILD_DIR) -archs $(3)
endef

gofmt:
	@GO111MODULE=off go fmt ./...

local: gofmt man
	$(call build,local, $(LOCAL_ARCHS), $(LOCAL_DOWNLOAD_ADDR))


testing: gofmt man
	$(call build,test, $(DEFAULT_ARCHS), $(TEST_DOWNLOAD_ADDR))


release: gofmt man
	$(call build,release, $(DEFAULT_ARCHS), $(RELEASE_DOWNLOAD_ADDR))


pub_local:
	$(call pub,local,$(LOCAL_DOWNLOAD_ADDR),$(LOCAL_ARCHS))


pub_testing:
	$(call pub,test,$(TEST_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))


pub_release:
	$(call pub,release,$(RELEASE_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

man:
	@packr2 clean
	@packr2

# local:
# 	$(call build,linux,amd64)
# 	@cp build/linux-amd64/$(BIN) /usr/local/security-checker/
