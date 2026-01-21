#!/bin/bash

# Tagger 构建脚本

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 获取当前 git tag，如果没有 tag 则使用 dev
VERSION=$(git describe --tags --exact-match 2>/dev/null || echo "dev")

# LDFLAGS 用于注入版本信息
LDFLAGS="-X github.com/AkaraChen/tagger/internal/config.Version=${VERSION}"

# 显示帮助信息
show_help() {
    echo "用法: ./build.sh [命令]"
    echo ""
    echo "命令:"
    echo "  build          构建开发版本 (默认)"
    echo "  release        构建发布版本 (需要 git tag)"
    echo "  install        安装到系统"
    echo "  clean          清理构建产物"
    echo "  test           运行测试"
    echo "  version        显示当前版本"
    echo "  help           显示此帮助信息"
}

# 构建开发版本
build() {
    echo -e "${BLUE}构建版本: ${VERSION}${NC}"
    go build -ldflags "${LDFLAGS}" -o tagger
    echo -e "${GREEN}✓ 构建成功: ./tagger${NC}"
}

# 构建发布版本
build_release() {
    if [ "${VERSION}" = "dev" ]; then
        echo -e "${RED}错误: 无法构建发布版本，需要先创建 git tag${NC}"
        echo "请先创建 tag: git tag v1.0.0"
        exit 1
    fi
    echo -e "${BLUE}构建发布版本: ${VERSION}${NC}"
    go build -ldflags "${LDFLAGS} -s -w" -o tagger
    echo -e "${GREEN}✓ 发布版本构建成功: ./tagger${NC}"
}

# 安装到系统
install() {
    echo -e "${BLUE}安装 tagger (版本: ${VERSION})${NC}"
    go install -ldflags "${LDFLAGS}"
    echo -e "${GREEN}✓ 已安装到 \$GOPATH/bin/tagger${NC}"
}

# 清理构建产物
clean() {
    echo -e "${BLUE}清理构建产物...${NC}"
    rm -f tagger
    echo -e "${GREEN}✓ 清理完成${NC}"
}

# 运行测试
run_test() {
    echo -e "${BLUE}运行测试...${NC}"
    go test ./...
    echo -e "${GREEN}✓ 测试完成${NC}"
}

# 显示版本
show_version() {
    echo "Version: ${VERSION}"
}

# 主逻辑
case "${1:-build}" in
    build)
        build
        ;;
    release)
        build_release
        ;;
    install)
        install
        ;;
    clean)
        clean
        ;;
    test)
        run_test
        ;;
    version)
        show_version
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}错误: 未知命令 '${1}'${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac
