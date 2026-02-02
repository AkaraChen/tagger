package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldOpenActionPage(t *testing.T) {
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
			name:     "nil github config",
			config:   &Config{GitHostingProvider: GitHub},
			expected: true,
		},
		{
			name:     "nil openActionPage",
			config:   &Config{GitHostingProvider: GitHub, GitHub: &GitHubConfig{}},
			expected: true,
		},
		{
			name: "openActionPage true",
			config: &Config{
				GitHostingProvider: GitHub,
				GitHub:             &GitHubConfig{OpenActionPage: boolPtr(true)},
			},
			expected: true,
		},
		{
			name: "openActionPage false",
			config: &Config{
				GitHostingProvider: GitHub,
				GitHub:             &GitHubConfig{OpenActionPage: boolPtr(false)},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldOpenActionPage()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsGitHub(t *testing.T) {
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
			name:     "GitHub provider",
			config:   &Config{GitHostingProvider: GitHub},
			expected: true,
		},
		{
			name:     "Other provider",
			config:   &Config{GitHostingProvider: Other},
			expected: false,
		},
		{
			name:     "empty provider",
			config:   &Config{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsGitHub()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsGitLab(t *testing.T) {
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
			name:     "GitLab provider",
			config:   &Config{GitHostingProvider: GitLab},
			expected: true,
		},
		{
			name:     "GitHub provider",
			config:   &Config{GitHostingProvider: GitHub},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsGitLab()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsGitea(t *testing.T) {
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
			name:     "Gitea provider",
			config:   &Config{GitHostingProvider: Gitea},
			expected: true,
		},
		{
			name:     "GitHub provider",
			config:   &Config{GitHostingProvider: GitHub},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsGitea()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetTagPrefix(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: "v",
		},
		{
			name:     "nil versioning",
			config:   &Config{},
			expected: "v",
		},
		{
			name:     "empty prefix",
			config:   &Config{Versioning: &VersioningConfig{TagPrefix: ""}},
			expected: "v",
		},
		{
			name:     "custom prefix",
			config:   &Config{Versioning: &VersioningConfig{TagPrefix: "release-"}},
			expected: "release-",
		},
		{
			name:     "no prefix",
			config:   &Config{Versioning: &VersioningConfig{TagPrefix: " "}},
			expected: " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetTagPrefix()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetDefaultPush(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected *bool
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: nil,
		},
		{
			name:     "nil defaults",
			config:   &Config{},
			expected: nil,
		},
		{
			name:     "nil push",
			config:   &Config{Defaults: &DefaultsConfig{}},
			expected: nil,
		},
		{
			name:     "push true",
			config:   &Config{Defaults: &DefaultsConfig{Push: boolPtr(true)}},
			expected: boolPtr(true),
		},
		{
			name:     "push false",
			config:   &Config{Defaults: &DefaultsConfig{Push: boolPtr(false)}},
			expected: boolPtr(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetDefaultPush()
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", *result)
				}
			} else {
				if result == nil {
					t.Errorf("expected %v, got nil", *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("expected %v, got %v", *tt.expected, *result)
				}
			}
		})
	}
}

func TestShouldSkipConfirmation(t *testing.T) {
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
			name:     "nil defaults",
			config:   &Config{},
			expected: false,
		},
		{
			name:     "skip false",
			config:   &Config{Defaults: &DefaultsConfig{SkipConfirmation: false}},
			expected: false,
		},
		{
			name:     "skip true",
			config:   &Config{Defaults: &DefaultsConfig{SkipConfirmation: true}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldSkipConfirmation()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetMessageTemplate(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: "",
		},
		{
			name:     "nil defaults",
			config:   &Config{},
			expected: "",
		},
		{
			name:     "empty template",
			config:   &Config{Defaults: &DefaultsConfig{MessageTemplate: ""}},
			expected: "",
		},
		{
			name:     "custom template",
			config:   &Config{Defaults: &DefaultsConfig{MessageTemplate: "Release {version}"}},
			expected: "Release {version}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetMessageTemplate()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestShouldOpenPipelinePage(t *testing.T) {
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
			name:     "nil gitlab config",
			config:   &Config{GitHostingProvider: GitLab},
			expected: true,
		},
		{
			name:     "nil openPipelinePage",
			config:   &Config{GitHostingProvider: GitLab, GitLab: &GitLabConfig{}},
			expected: true,
		},
		{
			name: "openPipelinePage true",
			config: &Config{
				GitHostingProvider: GitLab,
				GitLab:             &GitLabConfig{OpenPipelinePage: boolPtr(true)},
			},
			expected: true,
		},
		{
			name: "openPipelinePage false",
			config: &Config{
				GitHostingProvider: GitLab,
				GitLab:             &GitLabConfig{OpenPipelinePage: boolPtr(false)},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldOpenPipelinePage()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestShouldOpenGiteaActionsPage(t *testing.T) {
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
			name:     "nil gitea config",
			config:   &Config{GitHostingProvider: Gitea},
			expected: true,
		},
		{
			name:     "nil openActionsPage",
			config:   &Config{GitHostingProvider: Gitea, Gitea: &GiteaConfig{}},
			expected: true,
		},
		{
			name: "openActionsPage true",
			config: &Config{
				GitHostingProvider: Gitea,
				Gitea:              &GiteaConfig{OpenActionsPage: boolPtr(true)},
			},
			expected: true,
		},
		{
			name: "openActionsPage false",
			config: &Config{
				GitHostingProvider: Gitea,
				Gitea:              &GiteaConfig{OpenActionsPage: boolPtr(false)},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldOpenGiteaActionsPage()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
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
			"gitHostingProvider": "GitHub",
			"github": {
				"openActionPage": true
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
		if config.GitHostingProvider != GitHub {
			t.Errorf("expected GitHub, got %s", config.GitHostingProvider)
		}
		if !config.ShouldOpenActionPage() {
			t.Error("expected ShouldOpenActionPage to be true")
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
		if config.GitHostingProvider != GitHub {
			t.Errorf("expected GitHub, got %s", config.GitHostingProvider)
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
