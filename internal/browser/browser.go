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

// matchesPlatformBase 检查仓库 URL 是否匹配配置的 base URL
func matchesPlatformBase(repoURL string, baseURL string) bool {
	if baseURL == "" {
		return false
	}

	parsedRepo, err := url.Parse(repoURL)
	if err != nil {
		return false
	}

	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	return parsedRepo.Hostname() == parsedBase.Hostname()
}

// detectPlatformFromURL 从 URL 自动检测平台类型
func detectPlatformFromURL(repoURL string) config.PlatformType {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return ""
	}

	hostname := strings.ToLower(parsedURL.Hostname())

	if hostname == "github.com" {
		return config.PlatformGitHub
	}
	if hostname == "gitlab.com" || hostname == "gitlab.cn" {
		return config.PlatformGitLab
	}
	if hostname == "gitea.com" || hostname == "codeberg.org" {
		return config.PlatformGitea
	}

	return ""
}

// getCIPageSuffix 根据平台类型返回 CI 页面后缀
func getCIPageSuffix(platformType config.PlatformType) string {
	switch platformType {
	case config.PlatformGitHub:
		return "/actions"
	case config.PlatformGitLab:
		return "/-/pipelines"
	case config.PlatformGitea:
		return "/actions"
	default:
		return ""
	}
}

// getPlatformName 获取平台显示名称
func getPlatformName(platformType config.PlatformType) string {
	switch platformType {
	case config.PlatformGitHub:
		return "GitHub"
	case config.PlatformGitLab:
		return "GitLab"
	case config.PlatformGitea:
		return "Gitea"
	default:
		return "Unknown"
	}
}

// OpenRepository 处理打开仓库的逻辑
func OpenRepository(cfg *config.Config, gitClient *git.GitClient) error {
	repoURL, err := gitClient.GetRemoteURL()
	if err != nil {
		return fmt.Errorf("failed to get repository URL: %w", err)
	}

	var shouldOpenRepo bool
	var targetURL string

	platformType := cfg.GetPlatformType()
	platformBase := cfg.GetPlatformBase()

	if platformType != "" {
		// 有配置的平台类型
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("ℹ Detected Git Hosting Provider: %s", getPlatformName(platformType))))

		// 检查仓库 URL 是否匹配配置
		if platformBase != "" {
			if !matchesPlatformBase(repoURL, platformBase) {
				fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("⚠ Warning: Repository URL does not match configured base: %s", platformBase)))
			}
		}

		targetURL = repoURL
		if cfg.ShouldOpenActionPage() {
			ciSuffix := getCIPageSuffix(platformType)
			if ciSuffix != "" {
				targetURL = repoURL + ciSuffix
				fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("ℹ Opening %s CI page (configured in tagger.config.json)", getPlatformName(platformType))))
			}
		} else {
			fmt.Println(ui.InfoStyle.Render("ℹ Opening repository homepage (configured in tagger.config.json)"))
		}

		shouldOpenRepo = true
	} else {
		// 没有配置，尝试自动检测
		detectedPlatform := detectPlatformFromURL(repoURL)

		confirmed, err := ui.ConfirmOpenRepo()
		if err != nil {
			if err.Error() == "cancelled" {
				return fmt.Errorf("cancelled")
			}
			return fmt.Errorf("failed to confirm open repo: %w", err)
		}

		shouldOpenRepo = confirmed
		targetURL = repoURL

		if detectedPlatform != "" {
			ciSuffix := getCIPageSuffix(detectedPlatform)
			if ciSuffix != "" {
				targetURL = repoURL + ciSuffix
			}
		}
	}

	if shouldOpenRepo {
		err = Open(targetURL)
		if err != nil {
			fmt.Println(ui.ErrorStyle.Render(fmt.Sprintf("✗ Failed to open browser: %v", err)))
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  Repository URL: %s", targetURL)))
			return fmt.Errorf("failed to open browser: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Opening %s in browser...", targetURL)))
	}

	return nil
}
