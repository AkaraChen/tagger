package cmd

import (
	"fmt"
	"sort"

	"github.com/AkaraChen/tagger/internal/git"
	"github.com/AkaraChen/tagger/internal/semver"
	"github.com/AkaraChen/tagger/internal/ui"
	semverlib "github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
)

var (
	historyLimit int
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "显示版本历史",
	Long:  `显示仓库中的语义化版本标签历史`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHistory(historyLimit)
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 10, "显示的版本数量")
}

func runHistory(limit int) error {
	// 1. 初始化
	gitClient := git.NewGitClient(".")
	versionMgr := semver.NewVersionManager()

	// 2. 检查是否在 git 仓库中
	isRepo, err := gitClient.IsGitRepository()
	if err != nil {
		return fmt.Errorf("failed to check git repository: %w", err)
	}
	if !isRepo {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	// 3. 获取所有 tags 及其日期
	tagInfos, err := gitClient.GetTagsWithDates()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	if len(tagInfos) == 0 {
		fmt.Println(ui.InfoStyle.Render("No tags found in this repository"))
		return nil
	}

	// 4. 过滤符合 semver 格式的 tags
	type versionInfo struct {
		version *semverlib.Version
		tagInfo git.TagInfo
	}

	var validVersions []versionInfo

	for _, tagInfo := range tagInfos {
		versions, _ := versionMgr.ParseTags([]string{tagInfo.Name})
		if len(versions) > 0 {
			validVersions = append(validVersions, versionInfo{
				version: versions[0],
				tagInfo: tagInfo,
			})
		}
	}

	if len(validVersions) == 0 {
		fmt.Println(ui.InfoStyle.Render("No semantic version tags found in this repository"))
		fmt.Println(ui.HelpStyle.Render(fmt.Sprintf("Total tags: %d (none match vX.Y.Z format)", len(tagInfos))))
		return nil
	}

	// 5. 按版本号排序（从新到旧）
	sort.Slice(validVersions, func(i, j int) bool {
		return validVersions[i].version.GreaterThan(validVersions[j].version)
	})

	// 6. 限制显示数量
	if limit > 0 && limit < len(validVersions) {
		validVersions = validVersions[:limit]
	}

	// 7. 显示版本历史
	fmt.Println(ui.TitleStyle.Render("Version History"))
	fmt.Println()

	for i, vInfo := range validVersions {
		versionStr := versionMgr.FormatVersion(vInfo.version)
		dateStr := vInfo.tagInfo.Date.Format("2006-01-02")

		suffix := ""
		if i == 0 {
			suffix = ui.SuccessStyle.Render(" ← Latest")
		}

		fmt.Printf("%s  (%s)%s\n",
			ui.SelectedStyle.Render(versionStr),
			ui.HelpStyle.Render(dateStr),
			suffix,
		)
	}

	fmt.Println()
	if limit > 0 && limit < len(validVersions) {
		fmt.Println(ui.HelpStyle.Render(fmt.Sprintf("Showing %d of %d versions", limit, len(validVersions))))
	} else {
		fmt.Println(ui.HelpStyle.Render(fmt.Sprintf("Total: %d versions", len(validVersions))))
	}

	return nil
}
