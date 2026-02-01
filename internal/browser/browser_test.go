package browser

import (
	"testing"
)

func TestIsGitHub(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "github https url",
			url:      "https://github.com/user/repo",
			expected: true,
		},
		{
			name:     "github https url with .git",
			url:      "https://github.com/user/repo.git",
			expected: true,
		},
		{
			name:     "github ssh url",
			url:      "git@github.com:user/repo.git",
			expected: false, // ssh 格式解析后 hostname 为空或为不同格式
		},
		{
			name:     "github http url",
			url:      "http://github.com/user/repo",
			expected: true,
		},
		{
			name:     "github with www",
			url:      "https://www.github.com/user/repo",
			expected: false, // www.github.com != github.com
		},
		{
			name:     "github with subdomain",
			url:      "https://api.github.com/user/repo",
			expected: false,
		},
		{
			name:     "github with port",
			url:      "https://github.com:8443/user/repo",
			expected: true,
		},
		{
			name:     "gitlab url",
			url:      "https://gitlab.com/user/repo",
			expected: false,
		},
		{
			name:     "bitbucket url",
			url:      "https://bitbucket.org/user/repo",
			expected: false,
		},
		{
			name:     "self-hosted",
			url:      "https://git.example.com/user/repo",
			expected: false,
		},
		{
			name:     "invalid url",
			url:      "not a url",
			expected: false,
		},
		{
			name:     "empty url",
			url:      "",
			expected: false,
		},
		{
			name:     "github enterprise (not supported)",
			url:      "https://github.enterprise.com/user/repo",
			expected: false,
		},
		{
			name:     "path only",
			url:      "/user/repo",
			expected: false,
		},
		{
			name:     "github with uppercase",
			url:      "https://GITHUB.COM/user/repo",
			expected: false, // hostname 是大小写敏感的
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitHub(tt.url)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for url %s", tt.expected, result, tt.url)
			}
		})
	}
}
