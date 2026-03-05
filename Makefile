VERSION ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null | tr '/' '-')
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

BINARY  := cloudclaw
GOFLAGS := -trimpath

# 跨平台构建
GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: build test clean lint vet fmt run

build:
	go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/$(BINARY) ./cmd/cloudclaw

# 跨平台构建（用于发布）
build-all:
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/$(BINARY)-linux-amd64 ./cmd/cloudclaw
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/$(BINARY)-linux-arm64 ./cmd/cloudclaw
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/$(BINARY)-darwin-amd64 ./cmd/cloudclaw
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/$(BINARY)-darwin-arm64 ./cmd/cloudclaw

test:
	go test ./...

# 代码检查
lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

# 格式化并检查
check: fmt vet lint

# 运行
run:
	go run ./cmd/cloudclaw

# 清理
clean:
	rm -rf bin/
