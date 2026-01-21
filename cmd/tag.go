package cmd

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/AkaraChen/tagger/internal/config"
	"github.com/AkaraChen/tagger/internal/git"
	"github.com/AkaraChen/tagger/internal/semver"
	"github.com/AkaraChen/tagger/internal/ui"
)

// RunTag 执行 tag 创建命令
func RunTag(message string, autoPush, noPush, dryRun bool) error {
	// 1. 初始化
	gitClient := git.NewGitClient(".")
	versionMgr := semver.NewVersionManager()

	// 加载配置文件
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

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

			// 处理打开仓库的逻辑，优先使用配置文件
			if err := handleOpenRepository(cfg, gitClient); err != nil {
				// 打开仓库失败不应该影响整体流程，只输出错误信息
				if err.Error() != "cancelled" && err.Error() != "skipped" {
					fmt.Println(ui.ErrorStyle.Render(fmt.Sprintf("✗ %v", err)))
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

// isGitHub 判断仓库 URL 是否是 GitHub
// TODO: 暂不支持 GitHub Enterprise，仅支持 github.com
func isGitHub(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// 检查 hostname 是否为 github.com
	return parsedURL.Hostname() == "github.com"
}

// handleOpenRepository 处理打开仓库的逻辑，优先使用配置文件
func handleOpenRepository(cfg *config.Config, gitClient *git.GitClient) error {
	// 获取远程仓库 URL
	repoURL, err := gitClient.GetRemoteURL()
	if err != nil {
		return fmt.Errorf("failed to get repository URL: %w", err)
	}

	// 判断是否为 GitHub 仓库
	isGitHubRepo := isGitHub(repoURL)

	// 变量定义
	var shouldOpenRepo bool
	var targetURL string

	// 如果配置文件存在且指定了 gitHostingProvider
	if cfg != nil && cfg.GitHostingProvider != "" {
		// 显示检测到的配置信息
		providerName := string(cfg.GitHostingProvider)
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("ℹ Detected Git Hosting Provider: %s", providerName)))

		// 如果配置指定的是 GitHub
		if cfg.IsGitHub() {
			// 检查实际仓库是否为 GitHub
			if !isGitHubRepo {
				fmt.Println(ui.InfoStyle.Render("⚠ Warning: Config specifies GitHub, but repository URL is not github.com"))
			}

			// 根据配置决定目标 URL
			targetURL = repoURL
			if cfg.ShouldOpenActionPage() {
				targetURL = repoURL + "/actions"
				fmt.Println(ui.InfoStyle.Render("ℹ Opening GitHub Actions page (configured in tagger.config.json)"))
			} else {
				fmt.Println(ui.InfoStyle.Render("ℹ Opening repository homepage (configured in tagger.config.json)"))
			}

			shouldOpenRepo = true
		} else {
			// 配置指定为 Other，使用默认行为（询问用户）
			confirmed, err := ui.ConfirmOpenRepo()
			if err != nil {
				if err.Error() == "cancelled" {
					return fmt.Errorf("cancelled")
				}
				return fmt.Errorf("failed to confirm open repo: %w", err)
			}

			shouldOpenRepo = confirmed
			targetURL = repoURL
		}
	} else {
		// 没有配置文件，使用原有的交互逻辑
		confirmed, err := ui.ConfirmOpenRepo()
		if err != nil {
			if err.Error() == "cancelled" {
				return fmt.Errorf("cancelled")
			}
			return fmt.Errorf("failed to confirm open repo: %w", err)
		}

		shouldOpenRepo = confirmed

		// 根据实际 URL 判断是否为 GitHub
		targetURL = repoURL
		if isGitHubRepo {
			targetURL = repoURL + "/actions"
		}
	}

	// 如果确定要打开仓库
	if shouldOpenRepo {
		err = openBrowser(targetURL)
		if err != nil {
			fmt.Println(ui.ErrorStyle.Render(fmt.Sprintf("✗ Failed to open browser: %v", err)))
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  Repository URL: %s", targetURL)))
			return fmt.Errorf("failed to open browser: %w", err)
		}

		// 根据是否为 GitHub 和是否为 Actions 页面输出不同的成功信息
		if isGitHubRepo && targetURL == repoURL+"/actions" {
			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Opening GitHub Actions: %s", targetURL)))
		} else {
			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Opening %s in browser...", targetURL)))
		}
	}

	return nil
}
