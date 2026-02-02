package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ConfigFileName = "tagger.config.json"
const SchemaURL = "https://raw.githubusercontent.com/AkaraChen/tagger/main/tagger.schema.json"

// GitHostingProvider 表示 Git 托管平台类型
type GitHostingProvider string

const (
	GitHub GitHostingProvider = "GitHub"
	GitLab GitHostingProvider = "GitLab"
	Gitea  GitHostingProvider = "Gitea"
	Other  GitHostingProvider = "Other"
)

// DefaultsConfig 默认行为配置
type DefaultsConfig struct {
	Push             *bool  `json:"push,omitempty"`
	MessageTemplate  string `json:"messageTemplate,omitempty"`
	SkipConfirmation bool   `json:"skipConfirmation,omitempty"`
}

// VersioningConfig 版本号配置
type VersioningConfig struct {
	TagPrefix             string `json:"tagPrefix,omitempty"`
	DefaultPreReleaseType string `json:"defaultPreReleaseType,omitempty"`
}

// GitHubConfig GitHub 平台的配置
type GitHubConfig struct {
	// 使用指针类型可以区分"未设置"和"false"
	OpenActionPage *bool `json:"openActionPage,omitempty"`
}

// GitLabConfig GitLab 平台的配置
type GitLabConfig struct {
	OpenPipelinePage *bool `json:"openPipelinePage,omitempty"`
}

// GiteaConfig Gitea 平台的配置
type GiteaConfig struct {
	OpenActionsPage *bool `json:"openActionsPage,omitempty"`
}

// Config 工具的配置文件结构
type Config struct {
	Schema             string             `json:"$schema,omitempty"`
	GitHostingProvider GitHostingProvider `json:"gitHostingProvider"`
	Defaults           *DefaultsConfig    `json:"defaults,omitempty"`
	Versioning         *VersioningConfig  `json:"versioning,omitempty"`
	GitHub             *GitHubConfig      `json:"github,omitempty"`
	GitLab             *GitLabConfig      `json:"gitlab,omitempty"`
	Gitea              *GiteaConfig       `json:"gitea,omitempty"`
}

// Load 从当前目录加载配置文件
func Load() (*Config, error) {
	configPath := filepath.Join(".", ConfigFileName)

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // 配置文件不存在，返回 nil
	}

	// 读取文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// 解析 JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ShouldOpenActionPage 判断是否应该打开 Action 页面
// 如果配置中没有指定，默认返回 true
func (c *Config) ShouldOpenActionPage() bool {
	if c == nil || c.GitHub == nil || c.GitHub.OpenActionPage == nil {
		return true // 默认打开 Action 页面
	}
	return *c.GitHub.OpenActionPage
}

// IsGitHub 判断配置中的托管平台是否为 GitHub
func (c *Config) IsGitHub() bool {
	if c == nil {
		return false
	}
	return c.GitHostingProvider == GitHub
}

// IsGitLab 判断配置中的托管平台是否为 GitLab
func (c *Config) IsGitLab() bool {
	if c == nil {
		return false
	}
	return c.GitHostingProvider == GitLab
}

// IsGitea 判断配置中的托管平台是否为 Gitea
func (c *Config) IsGitea() bool {
	if c == nil {
		return false
	}
	return c.GitHostingProvider == Gitea
}

// ShouldOpenPipelinePage 判断是否应该打开 GitLab Pipeline 页面
func (c *Config) ShouldOpenPipelinePage() bool {
	if c == nil || c.GitLab == nil || c.GitLab.OpenPipelinePage == nil {
		return true
	}
	return *c.GitLab.OpenPipelinePage
}

// ShouldOpenGiteaActionsPage 判断是否应该打开 Gitea Actions 页面
func (c *Config) ShouldOpenGiteaActionsPage() bool {
	if c == nil || c.Gitea == nil || c.Gitea.OpenActionsPage == nil {
		return true
	}
	return *c.Gitea.OpenActionsPage
}

// GetTagPrefix 获取 tag 前缀，默认返回 "v"
func (c *Config) GetTagPrefix() string {
	if c == nil || c.Versioning == nil || c.Versioning.TagPrefix == "" {
		return "v"
	}
	return c.Versioning.TagPrefix
}

// GetDefaultPreReleaseType 获取默认预发布类型
func (c *Config) GetDefaultPreReleaseType() string {
	if c == nil || c.Versioning == nil {
		return ""
	}
	return c.Versioning.DefaultPreReleaseType
}

// GetDefaultPush 获取默认推送行为
// 返回 nil 表示询问用户，true 表示自动推送，false 表示不推送
func (c *Config) GetDefaultPush() *bool {
	if c == nil || c.Defaults == nil {
		return nil
	}
	return c.Defaults.Push
}

// ShouldSkipConfirmation 是否跳过创建确认
func (c *Config) ShouldSkipConfirmation() bool {
	if c == nil || c.Defaults == nil {
		return false
	}
	return c.Defaults.SkipConfirmation
}

// GetMessageTemplate 获取消息模板
func (c *Config) GetMessageTemplate() string {
	if c == nil || c.Defaults == nil {
		return ""
	}
	return c.Defaults.MessageTemplate
}

// CreateDefault 创建默认配置文件
func CreateDefault() error {
	// 检查文件是否已存在
	configPath := filepath.Join(".", ConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", configPath)
	}

	// 创建默认配置
	openActionPage := true
	config := Config{
		Schema:             SchemaURL,
		GitHostingProvider: GitHub,
		GitHub: &GitHubConfig{
			OpenActionPage: &openActionPage,
		},
	}

	// 序列化为 JSON（带缩进）
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
