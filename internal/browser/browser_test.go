package browser

import (
	"testing"

	"github.com/AkaraChen/tagger/internal/config"
)

func TestDetectPlatformFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected config.PlatformType
	}{
		{
			name:     "github https url",
			url:      "https://github.com/user/repo",
			expected: config.PlatformGitHub,
		},
		{
			name:     "github https url with .git",
			url:      "https://github.com/user/repo.git",
			expected: config.PlatformGitHub,
		},
		{
			name:     "github ssh url",
			url:      "git@github.com:user/repo.git",
			expected: "", // ssh 格式解析后 hostname 为空或为不同格式
		},
		{
			name:     "github http url",
			url:      "http://github.com/user/repo",
			expected: config.PlatformGitHub,
		},
		{
			name:     "github with www",
			url:      "https://www.github.com/user/repo",
			expected: "", // www.github.com != github.com
		},
		{
			name:     "github with subdomain",
			url:      "https://api.github.com/user/repo",
			expected: "",
		},
		{
			name:     "github with port",
			url:      "https://github.com:8443/user/repo",
			expected: config.PlatformGitHub,
		},
		{
			name:     "gitlab url",
			url:      "https://gitlab.com/user/repo",
			expected: config.PlatformGitLab,
		},
		{
			name:     "gitlab.cn url",
			url:      "https://gitlab.cn/user/repo",
			expected: config.PlatformGitLab,
		},
		{
			name:     "gitea url",
			url:      "https://gitea.com/user/repo",
			expected: config.PlatformGitea,
		},
		{
			name:     "codeberg url",
			url:      "https://codeberg.org/user/repo",
			expected: config.PlatformGitea,
		},
		{
			name:     "bitbucket url",
			url:      "https://bitbucket.org/user/repo",
			expected: "",
		},
		{
			name:     "self-hosted",
			url:      "https://git.example.com/user/repo",
			expected: "",
		},
		{
			name:     "invalid url",
			url:      "not a url",
			expected: "",
		},
		{
			name:     "empty url",
			url:      "",
			expected: "",
		},
		{
			name:     "github enterprise (not auto-detected)",
			url:      "https://github.enterprise.com/user/repo",
			expected: "",
		},
		{
			name:     "path only",
			url:      "/user/repo",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectPlatformFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for url %s", tt.expected, result, tt.url)
			}
		})
	}
}

func TestMatchesPlatformBase(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		baseURL  string
		expected bool
	}{
		{
			name:     "matching github enterprise",
			repoURL:  "https://github.enterprise.com/user/repo",
			baseURL:  "https://github.enterprise.com",
			expected: true,
		},
		{
			name:     "matching self-hosted gitlab",
			repoURL:  "https://gitlab.mycompany.com/user/repo",
			baseURL:  "https://gitlab.mycompany.com",
			expected: true,
		},
		{
			name:     "non-matching",
			repoURL:  "https://github.com/user/repo",
			baseURL:  "https://github.enterprise.com",
			expected: false,
		},
		{
			name:     "empty base url",
			repoURL:  "https://github.com/user/repo",
			baseURL:  "",
			expected: false,
		},
		{
			name:     "invalid repo url",
			repoURL:  "not a url",
			baseURL:  "https://github.com",
			expected: false,
		},
		{
			name:     "invalid base url",
			repoURL:  "https://github.com/user/repo",
			baseURL:  "not a url",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPlatformBase(tt.repoURL, tt.baseURL)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetCIPageSuffix(t *testing.T) {
	tests := []struct {
		name         string
		platformType config.PlatformType
		expected     string
	}{
		{
			name:         "github",
			platformType: config.PlatformGitHub,
			expected:     "/actions",
		},
		{
			name:         "gitlab",
			platformType: config.PlatformGitLab,
			expected:     "/-/pipelines",
		},
		{
			name:         "gitea",
			platformType: config.PlatformGitea,
			expected:     "/actions",
		},
		{
			name:         "unknown",
			platformType: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCIPageSuffix(tt.platformType)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetPlatformName(t *testing.T) {
	tests := []struct {
		name         string
		platformType config.PlatformType
		expected     string
	}{
		{
			name:         "github",
			platformType: config.PlatformGitHub,
			expected:     "GitHub",
		},
		{
			name:         "gitlab",
			platformType: config.PlatformGitLab,
			expected:     "GitLab",
		},
		{
			name:         "gitea",
			platformType: config.PlatformGitea,
			expected:     "Gitea",
		},
		{
			name:         "unknown",
			platformType: "",
			expected:     "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPlatformName(tt.platformType)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
