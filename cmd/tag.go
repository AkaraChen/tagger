package cmd

import (
	"fmt"
	"strings"

	"github.com/AkaraChen/tagger/internal/browser"
	"github.com/AkaraChen/tagger/internal/config"
	"github.com/AkaraChen/tagger/internal/git"
	"github.com/AkaraChen/tagger/internal/semver"
	"github.com/AkaraChen/tagger/internal/ui"
	"github.com/AkaraChen/tagger/internal/version"
)

// TagOptions 包含创建 tag 的所有选项
type TagOptions struct {
	Message        string
	AutoPush       bool
	NoPush         bool
	DryRun         bool
	PreReleaseType string
	PreReleaseNum  int
}

// RunTag 执行 tag 创建命令
func RunTag(opts TagOptions) error {
	gitClient := git.NewGitClient(".")
	versionMgr := semver.NewVersionManager()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	isRepo, err := gitClient.IsGitRepository()
	if err != nil {
		return fmt.Errorf("failed to check git repository: %w", err)
	}
	if !isRepo {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	hasChanges, err := gitClient.HasUncommittedChanges()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}
	if hasChanges {
		fmt.Println(ui.InfoStyle.Render("⚠ Warning: You have uncommitted changes"))
	}

	tags, err := gitClient.GetAllTags()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	versions, err := versionMgr.ParseTags(tags)
	if err != nil {
		return fmt.Errorf("failed to parse tags: %w", err)
	}

	// 获取 tag 前缀配置
	tagPrefix := cfg.GetTagPrefix()

	ctx := &tagContext{
		versionMgr:     versionMgr,
		selector:       version.NewSelectorWithPrefix(tagPrefix),
		versions:       versions,
		currentVersion: versionMgr.GetLatestVersion(versions),
		opts:           opts,
		tagPrefix:      tagPrefix,
	}

	ctx.displayVersionStatus()

	var newVersionStr string

	if versionMgr.IsPreRelease(ctx.currentVersion) {
		newVersionStr, err = ctx.handleExistingPreRelease()
	} else {
		newVersionStr, err = ctx.handleStableVersion()
	}

	if err != nil {
		if err.Error() == "cancelled" {
			fmt.Println(ui.InfoStyle.Render("Operation cancelled"))
			return nil
		}
		return err
	}

	tagMessage := opts.Message
	if tagMessage == "" {
		// 检查是否有配置的消息模板
		messageTemplate := cfg.GetMessageTemplate()
		if messageTemplate != "" {
			// 使用模板替换 {version} 占位符
			tagMessage = strings.ReplaceAll(messageTemplate, "{version}", newVersionStr)
		} else {
			addMessage, err := ui.ConfirmAddMessage()
			if err != nil {
				if err.Error() == "cancelled" {
					fmt.Println(ui.InfoStyle.Render("Operation cancelled"))
					return nil
				}
				return fmt.Errorf("failed to confirm add message: %w", err)
			}

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
	}

	// 检查是否应该跳过确认
	confirmed := true
	if !cfg.ShouldSkipConfirmation() {
		confirmed, err = ui.ConfirmCreateTag(ctx.currentVersionStr(), newVersionStr, tagMessage)
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
	}

	exists, err := gitClient.TagExists(newVersionStr)
	if err != nil {
		return fmt.Errorf("failed to check tag existence: %w", err)
	}
	if exists {
		return fmt.Errorf("tag %s already exists", newVersionStr)
	}

	if opts.DryRun {
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

	hasRemote, err := gitClient.HasRemote()
	if err != nil {
		return fmt.Errorf("failed to check remote: %w", err)
	}

	if !hasRemote {
		fmt.Println(ui.InfoStyle.Render("No remote repository configured, skipping push"))
		return nil
	}

	shouldPush := false

	if opts.AutoPush {
		shouldPush = true
	} else if opts.NoPush {
		shouldPush = false
	} else {
		// 检查配置中的默认推送行为
		defaultPush := cfg.GetDefaultPush()
		if defaultPush != nil {
			shouldPush = *defaultPush
		} else {
			// 没有配置默认值，询问用户
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
	}

	if shouldPush {
		if opts.DryRun {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("🔍 Dry run: Would push tag %s to remote", newVersionStr)))
		} else {
			fmt.Print(ui.InfoStyle.Render("⠋ Pushing tag to remote..."))
			err = gitClient.PushTag(newVersionStr)
			fmt.Print("\r")

			if err != nil {
				fmt.Println(ui.ErrorStyle.Render(fmt.Sprintf("✗ Failed to push tag: %v", err)))
				fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  You can manually push with: git push origin %s", newVersionStr)))
				return nil
			}

			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Tag %s pushed to remote successfully!", newVersionStr)))

			if err := browser.OpenRepository(cfg, gitClient); err != nil {
				if err.Error() != "cancelled" && err.Error() != "skipped" {
					fmt.Println(ui.ErrorStyle.Render(fmt.Sprintf("✗ %v", err)))
				}
			}
		}
	}

	return nil
}
