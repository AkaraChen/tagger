package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPlatformType(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected PlatformType
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: "",
		},
		{
			name:     "nil platform",
			config:   &Config{},
			expected: "",
		},
		{
			name:     "github platform",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitHub}},
			expected: PlatformGitHub,
		},
		{
			name:     "gitlab platform",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitLab}},
			expected: PlatformGitLab,
		},
		{
			name:     "gitea platform",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitea}},
			expected: PlatformGitea,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPlatformType()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetPlatformBase(t *testing.T) {
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
			name:     "nil platform",
			config:   &Config{},
			expected: "",
		},
		{
			name:     "empty base",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitHub}},
			expected: "",
		},
		{
			name:     "custom base",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitHub, Base: "https://github.example.com"}},
			expected: "https://github.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPlatformBase()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
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
			name:     "github platform",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitHub}},
			expected: true,
		},
		{
			name:     "gitlab platform",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitLab}},
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
			name:     "gitlab platform",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitLab}},
			expected: true,
		},
		{
			name:     "github platform",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitHub}},
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
			name:     "gitea platform",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitea}},
			expected: true,
		},
		{
			name:     "github platform",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitHub}},
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
			name:     "nil defaults",
			config:   &Config{},
			expected: true,
		},
		{
			name:     "nil openActionPage",
			config:   &Config{Defaults: &DefaultsConfig{}},
			expected: true,
		},
		{
			name:     "openActionPage true",
			config:   &Config{Defaults: &DefaultsConfig{OpenActionPage: boolPtr(true)}},
			expected: true,
		},
		{
			name:     "openActionPage false",
			config:   &Config{Defaults: &DefaultsConfig{OpenActionPage: boolPtr(false)}},
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

func TestGetTagTemplate(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: DefaultTagTemplate,
		},
		{
			name:     "nil versioning",
			config:   &Config{},
			expected: DefaultTagTemplate,
		},
		{
			name:     "empty template",
			config:   &Config{Versioning: &VersioningConfig{TagTemplate: ""}},
			expected: DefaultTagTemplate,
		},
		{
			name:     "custom template",
			config:   &Config{Versioning: &VersioningConfig{TagTemplate: "release-{major}.{minor}.{patch}"}},
			expected: "release-{major}.{minor}.{patch}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetTagTemplate()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetPreReleaseTagTemplate(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: DefaultPreReleaseTagTemplate,
		},
		{
			name:     "nil versioning",
			config:   &Config{},
			expected: DefaultPreReleaseTagTemplate,
		},
		{
			name:     "empty template",
			config:   &Config{Versioning: &VersioningConfig{PreReleaseTagTemplate: ""}},
			expected: DefaultPreReleaseTagTemplate,
		},
		{
			name:     "custom template",
			config:   &Config{Versioning: &VersioningConfig{PreReleaseTagTemplate: "v{major}.{minor}.{patch}-{preReleaseType}.{preReleaseNum}"}},
			expected: "v{major}.{minor}.{patch}-{preReleaseType}.{preReleaseNum}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPreReleaseTagTemplate()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetPreReleaseTypes(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: DefaultPreReleaseTypes,
		},
		{
			name:     "nil versioning",
			config:   &Config{},
			expected: DefaultPreReleaseTypes,
		},
		{
			name:     "empty types",
			config:   &Config{Versioning: &VersioningConfig{PreReleaseTypes: []string{}}},
			expected: DefaultPreReleaseTypes,
		},
		{
			name:     "custom types",
			config:   &Config{Versioning: &VersioningConfig{PreReleaseTypes: []string{"dev", "snapshot", "rc"}}},
			expected: []string{"dev", "snapshot", "rc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPreReleaseTypes()
			if len(result) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected %v, got %v", tt.expected, result)
					return
				}
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

func TestGetCIPageSuffix(t *testing.T) {
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
			name:     "github",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitHub}},
			expected: "/actions",
		},
		{
			name:     "gitlab",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitLab}},
			expected: "/-/pipelines",
		},
		{
			name:     "gitea",
			config:   &Config{Platform: &PlatformConfig{Type: PlatformGitea}},
			expected: "/actions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetCIPageSuffix()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
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
		if config.GetPlatformType() != PlatformGitHub {
			t.Errorf("expected github, got %s", config.GetPlatformType())
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
		if config.GetPlatformType() != PlatformGitHub {
			t.Errorf("expected github, got %s", config.GetPlatformType())
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
