package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/AkaraChen/tagger/internal/git"
	"github.com/AkaraChen/tagger/internal/semver"
	"github.com/AkaraChen/tagger/internal/ui"
)

// RunTag 执行 tag 创建命令
func RunTag(message string, autoPush, noPush, dryRun bool) error {
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

	// 3. 检查是否有未提交的修改（可选警告）
	hasChanges, err := gitClient.HasUncommittedChanges()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}
	if hasChanges {
		fmt.Println(ui.InfoStyle.Render("⚠ Warning: You have uncommitted changes"))
	}

	// 4. 获取所有 tags
	tags, err := gitClient.GetAllTags()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	// 5. 解析 tags，找到最新版本
	versions, err := versionMgr.ParseTags(tags)
	if err != nil {
		return fmt.Errorf("failed to parse tags: %w", err)
	}

	currentVersion := versionMgr.GetLatestVersion(versions)
	currentVersionStr := versionMgr.FormatVersion(currentVersion)

	// 计算所有可能的新版本（用于显示预览）
	patchVersion := versionMgr.FormatVersion(versionMgr.BumpPatch(currentVersion))
	minorVersion := versionMgr.FormatVersion(versionMgr.BumpMinor(currentVersion))
	majorVersion := versionMgr.FormatVersion(versionMgr.BumpMajor(currentVersion))

	// 6. 使用 Bubble Tea 选择更新类型
	bumpType, err := ui.SelectBumpType(currentVersionStr, patchVersion, minorVersion, majorVersion)
	if err != nil {
		if err.Error() == "cancelled" {
			fmt.Println(ui.InfoStyle.Render("Operation cancelled"))
			return nil
		}
		return fmt.Errorf("failed to select bump type: %w", err)
	}

	// 7. 计算新版本号
	newVersion, err := versionMgr.CalculateNewVersion(currentVersion, bumpType)
	if err != nil {
		return fmt.Errorf("failed to calculate new version: %w", err)
	}
	newVersionStr := versionMgr.FormatVersion(newVersion)

	// 8. 处理 tag message
	tagMessage := message
	if tagMessage == "" {
		// 询问是否添加 message
		addMessage, err := ui.ConfirmAddMessage()
		if err != nil {
			if err.Error() == "cancelled" {
				fmt.Println(ui.InfoStyle.Render("Operation cancelled"))
				return nil
			}
			return fmt.Errorf("failed to confirm add message: %w", err)
		}

		// 9. 如果用户选择添加 message，打开 textarea
		if addMessage {
			defaultText := fmt.Sprintf("Release %s: ", newVersionStr)
			tagMessage, err = ui.InputTagMessage(defaultText)
			if err != nil {
				if err.Error() == "cancelled" {
					fmt.Println(ui.InfoStyle.Render("Operation cancelled"))
					return nil
				}
				return fmt.Errorf("failed to input tag message: %w", err)
			}
		}
	}

	// 10. 确认创建 tag
	confirmed, err := ui.ConfirmCreateTag(currentVersionStr, newVersionStr, tagMessage)
	if err != nil {
		if err.Error() == "cancelled" {
			fmt.Println(ui.InfoStyle.Render("Operation cancelled"))
			return nil
		}
		return fmt.Errorf("failed to confirm create tag: %w", err)
	}

	if !confirmed {
		fmt.Println(ui.InfoStyle.Render("Operation cancelled"))
		return nil
	}

	// 11. 检查 tag 是否已存在
	exists, err := gitClient.TagExists(newVersionStr)
	if err != nil {
		return fmt.Errorf("failed to check tag existence: %w", err)
	}
	if exists {
		return fmt.Errorf("tag %s already exists", newVersionStr)
	}

	// 12. 创建 tag
	if dryRun {
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("🔍 Dry run: Would create tag %s", newVersionStr)))
		if tagMessage != "" {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   Message: %s", tagMessage)))
		}
	} else {
		if tagMessage != "" {
			err = gitClient.CreateAnnotatedTag(newVersionStr, tagMessage)
		} else {
			err = gitClient.CreateTag(newVersionStr)
		}

		if err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Tag %s created successfully!", newVersionStr)))
	}

	// 13. 检查是否有远程仓库
	hasRemote, err := gitClient.HasRemote()
	if err != nil {
		return fmt.Errorf("failed to check remote: %w", err)
	}

	if !hasRemote {
		fmt.Println(ui.InfoStyle.Render("No remote repository configured, skipping push"))
		return nil
	}

	// 14. 处理推送
	shouldPush := false

	if autoPush {
		shouldPush = true
	} else if !noPush {
		// 询问是否推送
		confirmed, err := ui.ConfirmPush(newVersionStr)
		if err != nil {
			if err.Error() == "cancelled" {
				fmt.Println(ui.InfoStyle.Render("Skipping push"))
				return nil
			}
			return fmt.Errorf("failed to confirm push: %w", err)
		}
		shouldPush = confirmed
	}

	// 15. 推送 tag
	if shouldPush {
		if dryRun {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("🔍 Dry run: Would push tag %s to remote", newVersionStr)))
		} else {
			fmt.Print(ui.InfoStyle.Render("⠋ Pushing tag to remote..."))
			err = gitClient.PushTag(newVersionStr)
			fmt.Print("\r") // 清除 spinner

			if err != nil {
				fmt.Println(ui.ErrorStyle.Render(fmt.Sprintf("✗ Failed to push tag: %v", err)))
				fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  You can manually push with: git push origin %s", newVersionStr)))
				return nil // 不返回错误，因为 tag 已经创建成功
			}

			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Tag %s pushed to remote successfully!", newVersionStr)))

			// 询问是否打开 GitHub 仓库
			shouldOpenRepo, err := ui.ConfirmOpenRepo()
			if err != nil && err.Error() != "cancelled" {
				return fmt.Errorf("failed to confirm open repo: %w", err)
			}

			if shouldOpenRepo {
				repoURL, err := gitClient.GetRemoteURL()
				if err != nil {
					fmt.Println(ui.ErrorStyle.Render(fmt.Sprintf("✗ Failed to get repository URL: %v", err)))
				} else {
					err = openBrowser(repoURL)
					if err != nil {
						fmt.Println(ui.ErrorStyle.Render(fmt.Sprintf("✗ Failed to open browser: %v", err)))
						fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  Repository URL: %s", repoURL)))
					} else {
						fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Opening %s in browser...", repoURL)))
					}
				}
			}
		}
	}

	return nil
}

// openBrowser 在默认浏览器中打开 URL
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
