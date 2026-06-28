APP     := sterm
MODULE  := github.com/ha1377311454/sterm
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
  -X '$(MODULE)/cmd.Version=$(VERSION)' \
  -X '$(MODULE)/cmd.Commit=$(COMMIT)'   \
  -X '$(MODULE)/cmd.BuildDate=$(DATE)'

GO      := go
GOBUILD := CGO_ENABLED=0 $(GO) build -trimpath -ldflags "$(LDFLAGS)"
DIST    := dist

PLATFORMS := \
  linux/amd64   \
  linux/arm64   \
  darwin/amd64  \
  darwin/arm64  \
  windows/amd64 \
  windows/arm64

.PHONY: all build install clean test lint help $(PLATFORMS)

## default: 为当前主机平台构建
all: build

## build: 为当前主机平台构建 → ./sterm
build:
	$(GOBUILD) -o $(APP) .

## install: 安装到 $GOPATH/bin
install:
	$(GO) install -trimpath -ldflags "$(LDFLAGS)" .

## release: 为所有平台交叉编译 → dist/*.exe（不打包）
release: $(PLATFORMS)
	@echo "  DONE"
	@ls -lh $(DIST)/

## linux/amd64、darwin/arm64、windows/amd64 … 各平台单独目标
$(PLATFORMS):
	$(eval OS   := $(word 1,$(subst /, ,$@)))
	$(eval ARCH := $(word 2,$(subst /, ,$@)))
	$(eval EXT  := $(if $(filter windows,$(OS)),.exe,))
	$(eval OUT  := $(DIST)/$(APP)_$(OS)_$(ARCH)$(EXT))
	@mkdir -p $(DIST)
	@echo "  BUILD   $(OUT)"
	@GOOS=$(OS) GOARCH=$(ARCH) $(GOBUILD) -o $(OUT) .

## test: 运行单元测试
test:
	$(GO) test ./... -v -count=1

## lint: 运行 golangci-lint（需单独安装）
lint:
	golangci-lint run ./...

## tidy: 整理 go.mod / go.sum
tidy:
	$(GO) mod tidy

## clean: 删除构建产物
clean:
	@rm -rf $(DIST) $(APP) $(APP).exe
	@echo "  CLEAN"

## version: 打印版本信息
version:
	@echo "version:    $(VERSION)"
	@echo "commit:     $(COMMIT)"
	@echo "build date: $(DATE)"

## help: 列出可用目标
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
	@echo ""
