package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AkaraChen/tagger/cmd"
	"github.com/AkaraChen/tagger/internal/config"
	"github.com/AkaraChen/tagger/internal/ui"
)

func main() {
	// 检查是否是 init 子命令
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runInitCommand()
		return
	}

	// 检查是否是 history 子命令
	if len(os.Args) > 1 && os.Args[1] == "history" {
		runHistoryCommand()
		return
	}

	// 默认 tag 命令的参数
	showVersion := flag.Bool("version", false, "显示版本信息")
	showVersionShort := flag.Bool("v", false, "显示版本信息（简写）")
	tagMessage := flag.String("message", "", "Tag 消息（创建 annotated tag）")
	tagMessageShort := flag.String("m", "", "Tag 消息（简写）")
	autoPush := flag.Bool("push", false, "自动推送到远程")
	noPush := flag.Bool("no-push", false, "不推送到远程")
	dryRun := flag.Bool("dry-run", false, "模拟运行")
	showHelp := flag.Bool("help", false, "显示帮助信息")
	showHelpShort := flag.Bool("h", false, "显示帮助信息（简写）")

	flag.Parse()

	// 显示版本
	if *showVersion || *showVersionShort {
		fmt.Printf("tagger version %s\n", config.GetVersion())
		os.Exit(0)
	}

	// 显示帮助
	if *showHelp || *showHelpShort {
		printHelp()
		os.Exit(0)
	}

	// 合并 -m 和 --message
	message := *tagMessage
	if message == "" {
		message = *tagMessageShort
	}

	// 运行 tag 命令
	if err := cmd.RunTag(message, *autoPush, *noPush, *dryRun); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func runInitCommand() {
	if err := cmd.RunInit(); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func runHistoryCommand() {
	// history 命令的参数
	historyCmd := flag.NewFlagSet("history", flag.ExitOnError)
	limit := historyCmd.Int("n", 10, "显示的版本数量")
	historyCmd.Parse(os.Args[2:])

	if err := cmd.RunHistory(*limit); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func printHelp() {
	help := `tagger - Git 语义化版本标签管理工具

用法:
  tagger [选项]              创建新的版本标签
  tagger init                创建配置文件
  tagger history [选项]      显示版本历史

Tag 命令选项:
  -m, --message <text>    Tag 消息（创建 annotated tag）
  --push                  自动推送到远程
  --no-push               不推送到远程
  --dry-run               模拟运行（不实际创建 tag）
  -v, --version           显示版本信息
  -h, --help              显示帮助信息

History 命令选项:
  -n <number>             显示的版本数量（默认: 10）

示例:
  tagger                                    # 交互式创建 tag
  tagger init                               # 创建配置文件 tagger.config.json
  tagger -m "Release notes"                 # 创建带消息的 tag
  tagger --push                             # 创建 tag 并自动推送
  tagger --dry-run                          # 模拟运行
  tagger history                            # 显示最近 10 个版本
  tagger history -n 20                      # 显示最近 20 个版本

更多信息: https://github.com/AkaraChen/tagger
`
	fmt.Print(help)
}
