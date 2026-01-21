package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const ConfigFileName = "tagger.config.json"

// GitHostingProvider 表示 Git 托管平台类型
type GitHostingProvider string

const (
	GitHub GitHostingProvider = "GitHub"
	Other  GitHostingProvider = "Other"
)

// GitHubConfig GitHub 平台的配置
type GitHubConfig struct {
	// 使用指针类型可以区分"未设置"和"false"
	OpenActionPage *bool `json:"openActionPage,omitempty"`
}

// Config 工具的配置文件结构
type Config struct {
	GitHostingProvider GitHostingProvider `json:"gitHostingProvider"`
	GitHub             *GitHubConfig      `json:"github,omitempty"`
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
