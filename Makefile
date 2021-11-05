.PHONY: build pub man

default: local

BIN = "scheck"
BUILD_DIR = build
PUB_DIR = pub
ENTRY = cmd/checker/main.go

# 正式环境
RELEASE_DOWNLOAD_ADDR = zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/security-checker

# 测试环境
TEST_DOWNLOAD_ADDR = zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker


LOCAL_ARCHS = "local"
LOCAL_DOWNLOAD_ADDR = $(LOCAL_OSS_BUCKET)"."$(LOCAL_OSS_HOST)"/"$(SC_USERNAME)"/"security-checker
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
	@sed "s,{{INSTALLER_BASE_URL}},$(3),g" install.sh.template > install.sh
	@sed "s,{{INSTALLER_BASE_URL}},$(3),g" install.ps1.template > install.ps1
	@GO111MODULE=off CGO_ENABLED=0 go run cmd/make/make.go -main $(ENTRY) -binary $(BIN) -build-dir $(BUILD_DIR) \
		 -env $(1) -pub-dir $(PUB_DIR) -archs $(2) -download-addr $(3)

endef

define pub
	@echo "publish $(1) ..."
	@echo "publish $(2) ..."
	@echo "publish $(3) ..."
	@GO111MODULE=off go run cmd/make/make.go -pub -env $(1) -pub-dir $(PUB_DIR) -binary $(BIN) -download-addr $(2) \
		-build-dir $(BUILD_DIR) -archs $(3)
endef

gofmt:
	@GO111MODULE=off go fmt ./...

local: deps
	$(call build,local,$(LOCAL_ARCHS),$(LOCAL_DOWNLOAD_ADDR))

local_all: deps
	$(call build,local,$(DEFAULT_ARCHS),$(LOCAL_DOWNLOAD_ADDR))

testing: deps
	$(call build,test,$(DEFAULT_ARCHS),$(TEST_DOWNLOAD_ADDR))

release: deps
	$(call build,release,$(DEFAULT_ARCHS),$(RELEASE_DOWNLOAD_ADDR))

pub_local:
	$(call pub,local,$(LOCAL_DOWNLOAD_ADDR),$(LOCAL_ARCHS))

pub_local_all:
	$(call pub,local,$(LOCAL_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

pub_testing:
	$(call pub,test,$(TEST_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

pub_release:
	$(call pub,release,$(RELEASE_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

man:
	@packr2 clean
	@packr2 --ignore-imports

vet: prepare
	@go vet ./...

prepare:
	@mkdir -p git
	@echo "$$GIT_INFO" > git/git.go

test: test_deps
	@truncate -s 0 test.output
	@echo "#####################" >> test.output
	@echo "#" $(DATE) >> test.output
	@echo "#" $(VERSION) >> test.output
	@echo "#####################" >> test.output
	for pkg in `go list ./...`; do \
		echo "# testing $$pkg..." >> test.output; \
		GO111MODULE=off CGO_ENABLED=0 go test -timeout 60s -cover -benchmem -bench . $$pkg |tee -a test.output; \
		echo "######################" >> test.output; \
	done

deps: man gofmt vet lint test
lint_deps: man gofmt vet
test_deps: man gofmt vet

lint: lint_deps
	@golangci-lint run --fix # https://golangci-lint.run/usage/install/#local-installation
