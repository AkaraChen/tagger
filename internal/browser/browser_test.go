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
