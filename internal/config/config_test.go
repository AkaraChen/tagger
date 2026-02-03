package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldOpenBrowser(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: true,
		},
		{
			name:     "nil defaults config",
			config:   &Config{},
			expected: true,
		},
		{
			name:     "nil openBrowser",
			config:   &Config{Defaults: &DefaultsConfig{}},
			expected: true,
		},
		{
			name: "openBrowser true",
			config: &Config{
				Defaults: &DefaultsConfig{OpenBrowser: boolPtr(true)},
			},
			expected: true,
		},
		{
			name: "openBrowser false",
			config: &Config{
				Defaults: &DefaultsConfig{OpenBrowser: boolPtr(false)},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldOpenBrowser()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestShouldAutoPush(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: false,
		},
		{
			name:     "nil defaults config",
			config:   &Config{},
			expected: false,
		},
		{
			name:     "nil autoPush",
			config:   &Config{Defaults: &DefaultsConfig{}},
			expected: false,
		},
		{
			name: "autoPush true",
			config: &Config{
				Defaults: &DefaultsConfig{AutoPush: boolPtr(true)},
			},
			expected: true,
		},
		{
			name: "autoPush false",
			config: &Config{
				Defaults: &DefaultsConfig{AutoPush: boolPtr(false)},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldAutoPush()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetectPlatformFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected PlatformType
	}{
		{
			name:     "GitHub HTTPS URL",
			url:      "https://github.com/user/repo.git",
			expected: PlatformGitHub,
		},
		{
			name:     "GitHub SSH URL",
			url:      "git@github.com:user/repo.git",
			expected: PlatformGitHub,
		},
		{
			name:     "GitLab HTTPS URL",
			url:      "https://gitlab.com/user/repo.git",
			expected: PlatformGitLab,
		},
		{
			name:     "GitLab SSH URL",
			url:      "git@gitlab.com:user/repo.git",
			expected: PlatformGitLab,
		},
		{
			name:     "Bitbucket HTTPS URL",
			url:      "https://bitbucket.org/user/repo.git",
			expected: PlatformBitbucket,
		},
		{
			name:     "Bitbucket SSH URL",
			url:      "git@bitbucket.org:user/repo.git",
			expected: PlatformBitbucket,
		},

		{
			name:     "Unknown platform",
			url:      "https://example.com/user/repo.git",
			expected: "",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectPlatformFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetPlatformType(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		originURL    string
		expectedType PlatformType
	}{
		{
			name: "config has platform type",
			config: &Config{
				Platform: &PlatformConfig{Type: PlatformGitLab},
			},
			originURL:    "https://github.com/user/repo.git",
			expectedType: PlatformGitLab,
		},
		{
			name:         "detect from origin URL",
			config:       &Config{},
			originURL:    "https://github.com/user/repo.git",
			expectedType: PlatformGitHub,
		},
		{
			name:         "nil config, detect from origin URL",
			config:       nil,
			originURL:    "https://gitlab.com/user/repo.git",
			expectedType: PlatformGitLab,
		},
		{
			name:         "unknown platform",
			config:       &Config{},
			originURL:    "https://example.com/repo.git",
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPlatformType(tt.originURL)
			if result != tt.expectedType {
				t.Errorf("expected %v, got %v", tt.expectedType, result)
			}
		})
	}
}

func TestGetPlatformBase(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		platformType PlatformType
		expectedBase string
	}{
		{
			name: "config has custom base",
			config: &Config{
				Platform: &PlatformConfig{Base: "https://custom.github.com"},
			},
			platformType: PlatformGitHub,
			expectedBase: "https://custom.github.com",
		},
		{
			name:         "GitHub default base",
			config:       &Config{},
			platformType: PlatformGitHub,
			expectedBase: "https://github.com",
		},
		{
			name:         "GitLab default base",
			config:       nil,
			platformType: PlatformGitLab,
			expectedBase: "https://gitlab.com",
		},
		{
			name:         "Bitbucket default base",
			config:       nil,
			platformType: PlatformBitbucket,
			expectedBase: "https://bitbucket.org",
		},
		{
			name:         "Gitea no default base",
			config:       nil,
			platformType: PlatformGitea,
			expectedBase: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPlatformBase(tt.platformType)
			if result != tt.expectedBase {
				t.Errorf("expected %v, got %v", tt.expectedBase, result)
			}
		})
	}
}

func TestGetActionsPath(t *testing.T) {
	tests := []struct {
		name         string
		platformType PlatformType
		expectedPath string
	}{
		{name: "GitHub", platformType: PlatformGitHub, expectedPath: "/actions"},
		{name: "GitLab", platformType: PlatformGitLab, expectedPath: "/-/pipelines"},
		{name: "Bitbucket", platformType: PlatformBitbucket, expectedPath: "/pipelines"},
		{name: "Gitea", platformType: PlatformGitea, expectedPath: "/actions"},
		{name: "Unknown", platformType: "", expectedPath: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetActionsPath(tt.platformType)
			if result != tt.expectedPath {
				t.Errorf("expected %v, got %v", tt.expectedPath, result)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Save current directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "tagger-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Chdir(tmpDir)

	t.Run("no config file", func(t *testing.T) {
		config, err := Load()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if config != nil {
			t.Error("expected nil config when file doesn't exist")
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		configContent := `{
			"$schema": "https://example.com/schema.json",
			"platform": {
				"type": "github"
			},
			"defaults": {
				"openBrowser": true,
				"autoPush": false
			}
		}`
		os.WriteFile(ConfigFileName, []byte(configContent), 0644)
		defer os.Remove(ConfigFileName)

		config, err := Load()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if config == nil {
			t.Fatal("expected config, got nil")
		}
		if config.Platform == nil || config.Platform.Type != PlatformGitHub {
			t.Errorf("expected github, got %v", config.Platform)
		}
		if !config.ShouldOpenBrowser() {
			t.Error("expected ShouldOpenBrowser to be true")
		}
		if config.ShouldAutoPush() {
			t.Error("expected ShouldAutoPush to be false")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		os.WriteFile(ConfigFileName, []byte("invalid json"), 0644)
		defer os.Remove(ConfigFileName)

		_, err := Load()
		if err == nil {
			t.Error("expected error for invalid json")
		}
	})
}

func TestCreateDefault(t *testing.T) {
	// Save current directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "tagger-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Chdir(tmpDir)

	t.Run("create new config", func(t *testing.T) {
		err := CreateDefault()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Verify file was created
		configPath := filepath.Join(".", ConfigFileName)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Load and verify content
		config, err := Load()
		if err != nil {
			t.Errorf("unexpected error loading config: %v", err)
		}
		if config.Defaults == nil {
			t.Fatal("expected defaults config, got nil")
		}
		if !config.ShouldOpenBrowser() {
			t.Error("expected ShouldOpenBrowser to be true by default")
		}
		if config.ShouldAutoPush() {
			t.Error("expected ShouldAutoPush to be false by default")
		}

		// Clean up
		os.Remove(ConfigFileName)
	})

	t.Run("config already exists", func(t *testing.T) {
		// Create a config file first
		os.WriteFile(ConfigFileName, []byte("{}"), 0644)
		defer os.Remove(ConfigFileName)

		err := CreateDefault()
		if err == nil {
			t.Error("expected error when config already exists")
		}
	})
}

func boolPtr(b bool) *bool {
	return &b
}
