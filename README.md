# tagger

> 一个简单易用的 Git 语义化版本标签管理工具

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ✨ 功能特性

- 🚀 **自动版本管理** - 自动检测当前最新版本，智能递增
- 📦 **语义化版本** - 完全支持 [Semantic Versioning](https://semver.org/)
- 💬 **现代化交互** - 使用 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 打造的精美 TUI
- 🔖 **灵活的标签** - 支持 lightweight 和 annotated tags
- 📤 **一键推送** - 可选择是否推送到远程仓库
- 📊 **版本历史** - 查看项目的版本演进历史
- ✨ **单二进制** - 无需额外依赖，开箱即用

## 📥 安装

### 方式 1: 使用 go install（推荐）

```bash
go install github.com/AkaraChen/tagger@latest
```

### 方式 2: 从源码构建

```bash
git clone https://github.com/AkaraChen/tagger.git
cd tagger
go build -o tagger
sudo mv tagger /usr/local/bin/
```

## 🚀 快速开始

在你的 Git 项目目录中运行：

```bash
tagger
```

工具会引导你完成以下步骤：

1. 📊 检测当前最新版本
2. 🎯 选择更新类型（Patch/Minor/Major）
3. 📝 可选添加 tag message
4. ✅ 确认创建标签
5. 📤 可选推送到远程

## 📖 使用方法

### 基本用法

```bash
# 交互式创建 tag
tagger

# 创建带消息的 annotated tag
tagger -m "Release v1.2.3: Added new features"

# 自动推送，不询问
tagger --push

# 不推送，不询问
tagger --no-push

# 模拟运行（查看会创建什么 tag，但不实际创建）
tagger --dry-run
```

### 查看版本历史

```bash
# 显示最近 10 个版本
tagger history

# 显示最近 20 个版本
tagger history -n 20
```

### 命令行选项

#### Tag 命令

```
-m, --message <text>    Tag 消息（创建 annotated tag）
--push                  自动推送到远程
--no-push               不推送到远程
--dry-run               模拟运行
-v, --version           显示版本信息
-h, --help              显示帮助信息
```

#### History 命令

```
-n <number>             显示的版本数量（默认: 10）
```

## 💡 使用示例

### 创建 Patch 版本（v1.2.3 → v1.2.4）

```bash
$ tagger
Current Version: v1.2.3
❯ Patch  (v1.2.3 → v1.2.4)
  Minor  (v1.2.3 → v1.3.0)
  Major  (v1.2.3 → v2.0.0)

Add a tag message? [y/N] n
Create tag v1.2.3 → v1.2.4? [Y/n] y
✓ Tag v1.2.4 created successfully!
```

### 创建带消息的 Minor 版本

```bash
$ tagger -m "Release v1.3.0: Major improvements"
# 选择 Minor
✓ Tag v1.3.0 created successfully!
Push tag v1.3.0 to remote? [Y/n] y
✓ Tag v1.3.0 pushed to remote successfully!
```

### 查看版本历史

```bash
$ tagger history -n 5
Version History

v1.3.0  (2025-01-14) ← Latest
v1.2.4  (2025-01-13)
v1.2.3  (2025-01-10)
v1.2.2  (2025-01-05)
v1.2.1  (2025-01-01)

Total: 5 versions
```

## 🎯 工作原理

1. **扫描 Tags** - 扫描所有 Git 标签，识别符合 `vX.Y.Z` 格式的标签
2. **识别最新版本** - 使用语义化版本规则找到最新版本
3. **计算新版本** - 根据你的选择计算新版本号：
   - **Patch**: v1.2.3 → v1.2.4（补丁更新，bug 修复）
   - **Minor**: v1.2.3 → v1.3.0（小版本更新，新功能）
   - **Major**: v1.2.3 → v2.0.0（大版本更新，破坏性变更）
4. **创建标签** - 创建新的 Git 标签（lightweight 或 annotated）
5. **推送到远程** - 可选推送到远程仓库

## 🔧 项目结构

```
tagger/
├── cmd/                    # 命令实现
│   ├── tag.go             # Tag 创建命令
│   └── history.go         # History 命令
├── internal/
│   ├── git/               # Git 操作封装
│   ├── semver/            # 语义化版本管理
│   └── ui/                # Bubble Tea 交互界面
│       ├── prompt.go      # 交互组件
│       └── styles.go      # Lipgloss 样式
└── main.go                # 程序入口
```

## ❓ 常见问题

### 如果项目还没有任何标签怎么办？

Tagger 会从 `v0.0.0` 开始，你可以选择创建 `v0.0.1`、`v0.1.0` 或 `v1.0.0`。

### 项目中有不符合语义化版本的标签怎么办？

Tagger 会忽略它们，只处理符合 `vX.Y.Z` 格式的标签。

### 为什么必须使用 v 前缀？

这是 Go 生态系统的惯例，也符合大多数项目的最佳实践。

### Tag 创建成功但推送失败怎么办？

不用担心，tag 已经在本地创建成功。你可以稍后手动推送：

```bash
git push origin vX.Y.Z
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

[MIT License](LICENSE)

## 🙏 致谢

本项目使用了以下优秀的开源库：

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI 框架
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI 组件库
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - 样式库
- [semver](https://github.com/Masterminds/semver) - 语义化版本解析

---

Made with ❤️ by [AkaraChen](https://github.com/AkaraChen)
