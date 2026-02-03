package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ConfigFileName = "tagger.config.json"
const SchemaURL = "https://raw.githubusercontent.com/AkaraChen/tagger/main/tagger.schema.json"

// PlatformType 表示 Git 托管平台类型
type PlatformType string

const (
	PlatformGitHub    PlatformType = "github"
	PlatformGitLab    PlatformType = "gitlab"
	PlatformBitbucket PlatformType = "bitbucket"
	PlatformGitea     PlatformType = "gitea"
)

// PlatformConfig 平台配置
type PlatformConfig struct {
	Type PlatformType `json:"type"`
	Base string       `json:"base,omitempty"`
}

// DefaultsConfig 默认行为配置
type DefaultsConfig struct {
	OpenBrowser *bool `json:"openBrowser,omitempty"`
	AutoPush    *bool `json:"autoPush,omitempty"`
}

// Config 工具的配置文件结构
type Config struct {
	Schema   string          `json:"$schema,omitempty"`
	Platform *PlatformConfig `json:"platform,omitempty"`
	Defaults *DefaultsConfig `json:"defaults,omitempty"`
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

// ShouldOpenBrowser 判断是否应该在推送后打开浏览器
// 如果配置中没有指定，默认返回 true
func (c *Config) ShouldOpenBrowser() bool {
	if c == nil || c.Defaults == nil || c.Defaults.OpenBrowser == nil {
		return true // 默认打开浏览器
	}
	return *c.Defaults.OpenBrowser
}

// ShouldAutoPush 判断是否应该在创建 tag 后自动推送
// 如果配置中没有指定，默认返回 false（需要用户确认）
func (c *Config) ShouldAutoPush() bool {
	if c == nil || c.Defaults == nil || c.Defaults.AutoPush == nil {
		return false // 默认不自动推送
	}
	return *c.Defaults.AutoPush
}

// GetPlatformType 获取平台类型
// 如果配置了 platform.type，直接返回
// 如果没有配置，尝试从 origin URL 解析
// 如果都无法确定，返回空字符串
func (c *Config) GetPlatformType(originURL string) PlatformType {
	if c != nil && c.Platform != nil && c.Platform.Type != "" {
		return c.Platform.Type
	}
	return DetectPlatformFromURL(originURL)
}

// GetPlatformBase 获取平台基础 URL
// 如果配置了 platform.base，直接返回
// 如果没有配置，根据平台类型返回默认的 base URL
func (c *Config) GetPlatformBase(platformType PlatformType) string {
	if c != nil && c.Platform != nil && c.Platform.Base != "" {
		return c.Platform.Base
	}
	return GetDefaultBaseURL(platformType)
}

// DetectPlatformFromURL 从 Git URL 中检测平台类型
func DetectPlatformFromURL(rawURL string) PlatformType {
	// 转换 SSH URL 为 HTTPS URL 格式以便解析
	url := rawURL
	if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
		url = strings.Replace(url, ":", "/", 1)
		url = "https://" + url
	}

	// 移除 .git 后缀
	url = strings.TrimSuffix(url, ".git")

	// 解析 host
	host := ""
	if strings.HasPrefix(url, "https://") {
		parts := strings.SplitN(url[8:], "/", 2)
		if len(parts) > 0 {
			host = parts[0]
		}
	}

	switch host {
	case "github.com":
		return PlatformGitHub
	case "gitlab.com":
		return PlatformGitLab
	case "bitbucket.org":
		return PlatformBitbucket
	}

	return ""
}

// GetDefaultBaseURL 获取平台的默认基础 URL
func GetDefaultBaseURL(platformType PlatformType) string {
	switch platformType {
	case PlatformGitHub:
		return "https://github.com"
	case PlatformGitLab:
		return "https://gitlab.com"
	case PlatformBitbucket:
		return "https://bitbucket.org"
	default:
		return ""
	}
}

// GetActionsPath 获取平台的 actions/pipelines 路径
func GetActionsPath(platformType PlatformType) string {
	switch platformType {
	case PlatformGitHub:
		return "/actions"
	case PlatformGitLab:
		return "/-/pipelines"
	case PlatformBitbucket:
		return "/pipelines"
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
	openBrowser := true
	autoPush := false
	config := Config{
		Schema: SchemaURL,
		Defaults: &DefaultsConfig{
			OpenBrowser: &openBrowser,
			AutoPush:    &autoPush,
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
