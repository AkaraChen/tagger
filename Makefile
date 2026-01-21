.PHONY: build build-release install clean test

# 获取当前 git tag，如果没有 tag 则使用 dev
VERSION := $(shell git describe --tags --exact-match 2>/dev/null || echo "dev")

# 构建标志
LDFLAGS := -X github.com/AkaraChen/tagger/internal/config.Version=$(VERSION)

# 默认构建（开发版本）
build:
	go build -ldflags "$(LDFLAGS)" -o tagger

# 发布版本构建（需要在 tag 上）
build-release:
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "Error: Cannot build release version without a git tag"; \
		echo "Please create a tag first: git tag v1.0.0"; \
		exit 1; \
	fi
	go build -ldflags "$(LDFLAGS) -s -w" -o tagger

# 安装到系统
install:
	go install -ldflags "$(LDFLAGS)"

# 清理构建产物
clean:
	rm -f tagger

# 运行测试
test:
	go test ./...

# 显示当前版本
version:
	@echo "Version: $(VERSION)"
