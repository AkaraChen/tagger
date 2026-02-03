package browser

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/AkaraChen/tagger/internal/config"
	"github.com/AkaraChen/tagger/internal/git"
	"github.com/AkaraChen/tagger/internal/ui"
)

// Open 在默认浏览器中打开 URL
func Open(url string) error {
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

// OpenRepository 处理打开仓库的逻辑
func OpenRepository(cfg *config.Config, gitClient *git.GitClient) error {
	repoURL, err := gitClient.GetRemoteURL()
	if err != nil {
		return fmt.Errorf("failed to get repository URL: %w", err)
	}

	// 检测平台类型
	platformType := cfg.GetPlatformType(repoURL)

	var shouldOpenRepo bool
	var targetURL string

	// 检查配置中是否应该打开浏览器
	if !cfg.ShouldOpenBrowser() {
		return nil
	}

	if platformType != "" {
		providerName := string(platformType)
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("ℹ Detected Git Hosting Provider: %s", providerName)))

		// 构建目标 URL
		targetURL = repoURL
		actionsPath := config.GetActionsPath(platformType)
		if actionsPath != "" {
			targetURL = repoURL + actionsPath
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("ℹ Opening %s page (configured in tagger.config.json)", getActionsPageName(platformType))))
		} else {
			fmt.Println(ui.InfoStyle.Render("ℹ Opening repository homepage (configured in tagger.config.json)"))
		}

		shouldOpenRepo = true
	} else {
		// 未知平台，询问用户
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

	if shouldOpenRepo {
		err = Open(targetURL)
		if err != nil {
			fmt.Println(ui.ErrorStyle.Render(fmt.Sprintf("✗ Failed to open browser: %v", err)))
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  Repository URL: %s", targetURL)))
			return fmt.Errorf("failed to open browser: %w", err)
		}

		if isActionsURL(targetURL, platformType) {
			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Opening %s: %s", getActionsPageName(platformType), targetURL)))
		} else {
			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Opening %s in browser...", targetURL)))
		}
	}

	return nil
}

// isActionsURL 检查 URL 是否是 actions/pipelines 页面
func isActionsURL(targetURL string, platformType config.PlatformType) bool {
	actionsPath := config.GetActionsPath(platformType)
	if actionsPath == "" {
		return false
	}
	return strings.HasSuffix(targetURL, actionsPath)
}

// getActionsPageName 获取 Actions 页面的显示名称
func getActionsPageName(platformType config.PlatformType) string {
	switch platformType {
	case config.PlatformGitHub:
		return "GitHub Actions"
	case config.PlatformGitLab:
		return "GitLab CI/CD Pipelines"
	case config.PlatformBitbucket:
		return "Bitbucket Pipelines"
	case config.PlatformGitea:
		return "Gitea Actions"
	default:
		return "Actions"
	}
}

// IsGitHub 判断仓库 URL 是否是 GitHub（向后兼容）
func IsGitHub(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsedURL.Hostname() == "github.com"
}
