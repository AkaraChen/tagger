package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitClient 封装 Git 操作
type GitClient struct {
	workDir string
}

// TagInfo 包含 tag 的信息
type TagInfo struct {
	Name string
	Date time.Time
}

// NewGitClient 创建一个新的 GitClient
func NewGitClient(workDir string) *GitClient {
	return &GitClient{workDir: workDir}
}

// IsGitRepository 检查当前目录是否在 git 仓库中
func (g *GitClient) IsGitRepository() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.workDir
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

// HasUncommittedChanges 检查是否有未提交的修改
func (g *GitClient) HasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.workDir

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	return out.String() != "", nil
}

// GetAllTags 获取所有 tags
func (g *GitClient) GetAllTags() ([]string, error) {
	cmd := exec.Command("git", "tag", "-l")
	cmd.Dir = g.workDir

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		return []string{}, nil
	}

	tags := strings.Split(output, "\n")
	return tags, nil
}

// TagExists 检查指定的 tag 是否存在
func (g *GitClient) TagExists(tag string) (bool, error) {
	cmd := exec.Command("git", "tag", "-l", tag)
	cmd.Dir = g.workDir

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to check tag existence: %w", err)
	}

	return strings.TrimSpace(out.String()) != "", nil
}

// CreateTag 创建 lightweight tag
func (g *GitClient) CreateTag(version string) error {
	cmd := exec.Command("git", "tag", version)
	cmd.Dir = g.workDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tag: %s", stderr.String())
	}

	return nil
}

// CreateAnnotatedTag 创建 annotated tag
func (g *GitClient) CreateAnnotatedTag(version, message string) error {
	cmd := exec.Command("git", "tag", "-a", version, "-m", message)
	cmd.Dir = g.workDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create annotated tag: %s", stderr.String())
	}

	return nil
}

// HasRemote 检查是否配置了远程仓库
func (g *GitClient) HasRemote() (bool, error) {
	cmd := exec.Command("git", "remote")
	cmd.Dir = g.workDir

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false, err
	}

	return strings.TrimSpace(out.String()) != "", nil
}

// GetRemoteName 获取远程仓库名称（优先返回 origin）
func (g *GitClient) GetRemoteName() (string, error) {
	cmd := exec.Command("git", "remote")
	cmd.Dir = g.workDir

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", err
	}

	remotes := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(remotes) == 0 || remotes[0] == "" {
		return "", fmt.Errorf("no remote repository found")
	}

	// 优先使用 origin
	for _, remote := range remotes {
		if remote == "origin" {
			return "origin", nil
		}
	}

	return remotes[0], nil
}

// PushTag 推送 tag 到远程仓库
func (g *GitClient) PushTag(version string) error {
	// 获取远程名称
	remote, err := g.GetRemoteName()
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "push", remote, version)
	cmd.Dir = g.workDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push tag: %s", stderr.String())
	}

	return nil
}

// GetRemoteURL 获取远程仓库的 URL
func (g *GitClient) GetRemoteURL() (string, error) {
	remote, err := g.GetRemoteName()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "remote", "get-url", remote)
	cmd.Dir = g.workDir

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	url := strings.TrimSpace(out.String())

	// 转换 SSH URL 为 HTTPS URL
	// git@github.com:user/repo.git -> https://github.com/user/repo
	if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
		url = strings.Replace(url, ":", "/", 1)
		url = "https://" + url
	}

	// 移除 .git 后缀
	url = strings.TrimSuffix(url, ".git")

	return url, nil
}

// GetTagsWithDates 获取所有 tags 及其创建日期
func (g *GitClient) GetTagsWithDates() ([]TagInfo, error) {
	cmd := exec.Command("git", "for-each-ref", "--sort=-creatordate", "--format=%(refname:short)|%(creatordate:short)", "refs/tags")
	cmd.Dir = g.workDir

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get tags with dates: %w", err)
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		return []TagInfo{}, nil
	}

	lines := strings.Split(output, "\n")
	tagInfos := make([]TagInfo, 0, len(lines))

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}

		date, err := time.Parse("2006-01-02", parts[1])
		if err != nil {
			// 如果解析失败，使用零值时间
			date = time.Time{}
		}

		tagInfos = append(tagInfos, TagInfo{
			Name: parts[0],
			Date: date,
		})
	}

	return tagInfos, nil
}
