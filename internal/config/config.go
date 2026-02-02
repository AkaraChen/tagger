package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ConfigFileName = "tagger.config.json"
const SchemaURL = "https://raw.githubusercontent.com/AkaraChen/tagger/main/tagger.schema.json"

// 默认值常量
const (
	DefaultTagTemplate           = "v{major}.{minor}.{patch}"
	DefaultPreReleaseTagTemplate = "v{major}.{minor}.{patch}-{preReleaseType}{preReleaseNum}"
)

// DefaultPreReleaseTypes 默认预发布类型列表
var DefaultPreReleaseTypes = []string{"alpha", "beta", "rc"}

// PlatformType 平台类型
type PlatformType string

const (
	PlatformGitHub PlatformType = "github"
	PlatformGitLab PlatformType = "gitlab"
	PlatformGitea  PlatformType = "gitea"
)

// PlatformConfig 平台配置
type PlatformConfig struct {
	Type PlatformType `json:"type"`
	Base string       `json:"base,omitempty"`
}

// DefaultsConfig 默认行为配置
type DefaultsConfig struct {
	Push             *bool  `json:"push,omitempty"`
	MessageTemplate  string `json:"messageTemplate,omitempty"`
	SkipConfirmation bool   `json:"skipConfirmation,omitempty"`
	OpenActionPage   *bool  `json:"openActionPage,omitempty"`
}

// VersioningConfig 版本号配置
type VersioningConfig struct {
	TagTemplate           string   `json:"tagTemplate,omitempty"`
	PreReleaseTagTemplate string   `json:"preReleaseTagTemplate,omitempty"`
	PreReleaseTypes       []string `json:"preReleaseTypes,omitempty"`
}

// Config 工具的配置文件结构
type Config struct {
	Schema     string            `json:"$schema,omitempty"`
	Platform   *PlatformConfig   `json:"platform,omitempty"`
	Defaults   *DefaultsConfig   `json:"defaults,omitempty"`
	Versioning *VersioningConfig `json:"versioning,omitempty"`
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

// GetPlatformType 获取平台类型
func (c *Config) GetPlatformType() PlatformType {
	if c == nil || c.Platform == nil {
		return ""
	}
	return c.Platform.Type
}

// GetPlatformBase 获取平台 base URL
func (c *Config) GetPlatformBase() string {
	if c == nil || c.Platform == nil {
		return ""
	}
	return c.Platform.Base
}

// IsGitHub 判断配置中的托管平台是否为 GitHub
func (c *Config) IsGitHub() bool {
	return c.GetPlatformType() == PlatformGitHub
}

// IsGitLab 判断配置中的托管平台是否为 GitLab
func (c *Config) IsGitLab() bool {
	return c.GetPlatformType() == PlatformGitLab
}

// IsGitea 判断配置中的托管平台是否为 Gitea
func (c *Config) IsGitea() bool {
	return c.GetPlatformType() == PlatformGitea
}

// ShouldOpenActionPage 判断是否应该打开 CI 页面
func (c *Config) ShouldOpenActionPage() bool {
	if c == nil || c.Defaults == nil || c.Defaults.OpenActionPage == nil {
		return true // 默认打开
	}
	return *c.Defaults.OpenActionPage
}

// GetDefaultPush 获取默认推送行为
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

// GetTagTemplate 获取稳定版本 tag 模板
func (c *Config) GetTagTemplate() string {
	if c == nil || c.Versioning == nil || c.Versioning.TagTemplate == "" {
		return DefaultTagTemplate
	}
	return c.Versioning.TagTemplate
}

// GetPreReleaseTagTemplate 获取预发布版本 tag 模板
func (c *Config) GetPreReleaseTagTemplate() string {
	if c == nil || c.Versioning == nil || c.Versioning.PreReleaseTagTemplate == "" {
		return DefaultPreReleaseTagTemplate
	}
	return c.Versioning.PreReleaseTagTemplate
}

// GetPreReleaseTypes 获取预发布类型列表
func (c *Config) GetPreReleaseTypes() []string {
	if c == nil || c.Versioning == nil || len(c.Versioning.PreReleaseTypes) == 0 {
		return DefaultPreReleaseTypes
	}
	return c.Versioning.PreReleaseTypes
}

// GetCIPageSuffix 根据平台类型返回 CI 页面后缀
func (c *Config) GetCIPageSuffix() string {
	switch c.GetPlatformType() {
	case PlatformGitHub:
		return "/actions"
	case PlatformGitLab:
		return "/-/pipelines"
	case PlatformGitea:
		return "/actions"
	default:
		return ""
	}
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
		Schema: SchemaURL,
		Platform: &PlatformConfig{
			Type: PlatformGitHub,
		},
		Defaults: &DefaultsConfig{
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
