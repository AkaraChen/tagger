package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ConfigFileName = "tagger.config.json"
const SchemaURLTemplate = "https://raw.githubusercontent.com/AkaraChen/tagger/%s/tagger.schema.json"

// Version 在构建时通过 ldflags 注入
var Version = "dev"

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
	Schema             string             `json:"$schema,omitempty"`
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

// GetSchemaURL 获取 JSON Schema URL
func GetSchemaURL() string {
	version := Version
	if version == "dev" || version == "" {
		version = "main"
	}
	return fmt.Sprintf(SchemaURLTemplate, version)
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
		Schema:             GetSchemaURL(),
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
