package browser

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"

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

// IsGitHub 判断仓库 URL 是否是 GitHub
func IsGitHub(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsedURL.Hostname() == "github.com"
}

// OpenRepository 处理打开仓库的逻辑
func OpenRepository(cfg *config.Config, gitClient *git.GitClient) error {
	repoURL, err := gitClient.GetRemoteURL()
	if err != nil {
		return fmt.Errorf("failed to get repository URL: %w", err)
	}

	isGitHubRepo := IsGitHub(repoURL)

	var shouldOpenRepo bool
	var targetURL string

	if cfg != nil && cfg.GitHostingProvider != "" {
		providerName := string(cfg.GitHostingProvider)
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("ℹ Detected Git Hosting Provider: %s", providerName)))

		if cfg.IsGitHub() {
			if !isGitHubRepo {
				fmt.Println(ui.InfoStyle.Render("⚠ Warning: Config specifies GitHub, but repository URL is not github.com"))
			}

			targetURL = repoURL
			if cfg.ShouldOpenActionPage() {
				targetURL = repoURL + "/actions"
				fmt.Println(ui.InfoStyle.Render("ℹ Opening GitHub Actions page (configured in tagger.config.json)"))
			} else {
				fmt.Println(ui.InfoStyle.Render("ℹ Opening repository homepage (configured in tagger.config.json)"))
			}

			shouldOpenRepo = true
		} else {
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
		confirmed, err := ui.ConfirmOpenRepo()
		if err != nil {
			if err.Error() == "cancelled" {
				return fmt.Errorf("cancelled")
			}
			return fmt.Errorf("failed to confirm open repo: %w", err)
		}

		shouldOpenRepo = confirmed
		targetURL = repoURL
		if isGitHubRepo {
			targetURL = repoURL + "/actions"
		}
	}

	if shouldOpenRepo {
		err = Open(targetURL)
		if err != nil {
			fmt.Println(ui.ErrorStyle.Render(fmt.Sprintf("✗ Failed to open browser: %v", err)))
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  Repository URL: %s", targetURL)))
			return fmt.Errorf("failed to open browser: %w", err)
		}

		if isGitHubRepo && targetURL == repoURL+"/actions" {
			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Opening GitHub Actions: %s", targetURL)))
		} else {
			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Opening %s in browser...", targetURL)))
		}
	}

	return nil
}
