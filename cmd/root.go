package cmd

import (
	"fmt"
	"os"

	"github.com/AkaraChen/tagger/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Tag 命令参数
	tagMessage     string
	autoPush       bool
	noPush         bool
	dryRun         bool
	preReleaseType string
	preReleaseNum  int
)

// rootCmd 代表 tag 命令（默认命令）
var rootCmd = &cobra.Command{
	Use:   "tagger",
	Short: "Git 语义化版本标签管理工具",
	Long:  `Tagger 是一个用于创建和管理 Git 语义化版本标签的工具`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunTag(TagOptions{
			Message:        tagMessage,
			AutoPush:       autoPush,
			NoPush:         noPush,
			DryRun:         dryRun,
			PreReleaseType: preReleaseType,
			PreReleaseNum:  preReleaseNum,
		})
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&tagMessage, "message", "m", "", "Tag 消息（创建 annotated tag）")
	rootCmd.Flags().BoolVar(&autoPush, "push", false, "自动推送到远程")
	rootCmd.Flags().BoolVar(&noPush, "no-push", false, "不推送到远程")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "模拟运行")
	rootCmd.Flags().StringVar(&preReleaseType, "pre", "", "创建预发布版本 (alpha|beta|rc)")
	rootCmd.Flags().IntVar(&preReleaseNum, "pre-num", 0, "预发布版本号 (默认自动递增)")
}
