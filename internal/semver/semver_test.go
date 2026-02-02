package semver

import (
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestNewVersionManager(t *testing.T) {
	vm := NewVersionManager()
	if vm == nil {
		t.Fatal("NewVersionManager returned nil")
	}
}

func TestParseTags(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		name     string
		tags     []string
		expected int
	}{
		{
			name:     "valid semver tags",
			tags:     []string{"v1.0.0", "v1.1.0", "v2.0.0"},
			expected: 3,
		},
		{
			name:     "mixed valid and invalid tags",
			tags:     []string{"v1.0.0", "invalid", "v2.0.0", "not-a-version"},
			expected: 2,
		},
		{
			name:     "tags without v prefix",
			tags:     []string{"1.0.0", "2.0.0"},
			expected: 2,
		},
		{
			name:     "empty tags",
			tags:     []string{},
			expected: 0,
		},
		{
			name:     "pre-release tags",
			tags:     []string{"v1.0.0-alpha1", "v1.0.0-beta1", "v1.0.0-rc1"},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions, err := vm.ParseTags(tt.tags)
			if err != nil {
				t.Fatalf("ParseTags returned error: %v", err)
			}
			if len(versions) != tt.expected {
				t.Errorf("expected %d versions, got %d", tt.expected, len(versions))
			}
		})
	}
}

func TestGetLatestVersion(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{
			name:     "multiple versions",
			tags:     []string{"v1.0.0", "v2.0.0", "v1.5.0"},
			expected: "2.0.0",
		},
		{
			name:     "single version",
			tags:     []string{"v1.0.0"},
			expected: "1.0.0",
		},
		{
			name:     "empty versions returns 0.0.0",
			tags:     []string{},
			expected: "0.0.0",
		},
		{
			name:     "pre-release versions are less than stable",
			tags:     []string{"v1.0.0", "v1.0.0-alpha1"},
			expected: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions, _ := vm.ParseTags(tt.tags)
			latest := vm.GetLatestVersion(versions)
			if latest.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, latest.String())
			}
		})
	}
}

func TestBumpMajor(t *testing.T) {
	vm := NewVersionManager()
	v, _ := semver.NewVersion("1.2.3")

	result := vm.BumpMajor(v)
	if result.String() != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", result.String())
	}
}

func TestBumpMinor(t *testing.T) {
	vm := NewVersionManager()
	v, _ := semver.NewVersion("1.2.3")

	result := vm.BumpMinor(v)
	if result.String() != "1.3.0" {
		t.Errorf("expected 1.3.0, got %s", result.String())
	}
}

func TestBumpPatch(t *testing.T) {
	vm := NewVersionManager()
	v, _ := semver.NewVersion("1.2.3")

	result := vm.BumpPatch(v)
	if result.String() != "1.2.4" {
		t.Errorf("expected 1.2.4, got %s", result.String())
	}
}

func TestFormatVersion(t *testing.T) {
	vm := NewVersionManager()
	v, _ := semver.NewVersion("1.2.3")

	result := vm.FormatVersion(v)
	if result != "v1.2.3" {
		t.Errorf("expected v1.2.3, got %s", result)
	}
}

func TestCalculateNewVersion(t *testing.T) {
	vm := NewVersionManager()
	v, _ := semver.NewVersion("1.2.3")

	tests := []struct {
		bumpType string
		expected string
		hasError bool
	}{
		{"major", "2.0.0", false},
		{"minor", "1.3.0", false},
		{"patch", "1.2.4", false},
		{"MAJOR", "2.0.0", false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.bumpType, func(t *testing.T) {
			result, err := vm.CalculateNewVersion(v, tt.bumpType)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result.String() != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, result.String())
				}
			}
		})
	}
}

func TestIsPreRelease(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		version  string
		expected bool
	}{
		{"1.0.0", false},
		{"1.0.0-alpha1", true},
		{"1.0.0-beta2", true},
		{"1.0.0-rc1", true},
		{"2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, _ := semver.NewVersion(tt.version)
			result := vm.IsPreRelease(v)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetPreReleaseInfo(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		version      string
		expectedType PreReleaseType
		expectedNum  int
		isNil        bool
	}{
		{"1.0.0-alpha1", PreReleaseAlpha, 1, false},
		{"1.0.0-alpha5", PreReleaseAlpha, 5, false},
		{"1.0.0-beta2", PreReleaseBeta, 2, false},
		{"1.0.0-rc3", PreReleaseRC, 3, false},
		{"1.0.0", "", 0, true},
		{"1.0.0-invalid", "", 0, true},
		{"1.0.0-alpha", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, _ := semver.NewVersion(tt.version)
			info := vm.GetPreReleaseInfo(v)
			if tt.isNil {
				if info != nil {
					t.Errorf("expected nil, got %+v", info)
				}
			} else {
				if info == nil {
					t.Fatal("expected info, got nil")
				}
				if info.Type != tt.expectedType {
					t.Errorf("expected type %s, got %s", tt.expectedType, info.Type)
				}
				if info.Number != tt.expectedNum {
					t.Errorf("expected number %d, got %d", tt.expectedNum, info.Number)
				}
			}
		})
	}
}

func TestSetPreRelease(t *testing.T) {
	vm := NewVersionManager()
	v, _ := semver.NewVersion("1.2.3")

	tests := []struct {
		preType  PreReleaseType
		num      int
		expected string
	}{
		{PreReleaseAlpha, 1, "1.2.3-alpha1"},
		{PreReleaseBeta, 2, "1.2.3-beta2"},
		{PreReleaseRC, 3, "1.2.3-rc3"},
	}

	for _, tt := range tests {
		t.Run(string(tt.preType), func(t *testing.T) {
			result, err := vm.SetPreRelease(v, tt.preType, tt.num)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestBumpPreRelease(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		version  string
		expected string
		hasError bool
	}{
		{"1.0.0-alpha1", "1.0.0-alpha2", false},
		{"1.0.0-beta5", "1.0.0-beta6", false},
		{"1.0.0-rc1", "1.0.0-rc2", false},
		{"1.0.0", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, _ := semver.NewVersion(tt.version)
			result, err := vm.BumpPreRelease(v)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result.String() != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, result.String())
				}
			}
		})
	}
}

func TestNextPreReleaseStage(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		version  string
		expected string
		hasError bool
	}{
		{"1.0.0-alpha3", "1.0.0-beta1", false},
		{"1.0.0-beta2", "1.0.0-rc1", false},
		{"1.0.0-rc1", "", true},
		{"1.0.0", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, _ := semver.NewVersion(tt.version)
			result, err := vm.NextPreReleaseStage(v)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result.String() != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, result.String())
				}
			}
		})
	}
}

func TestReleaseStable(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		version  string
		expected string
		hasError bool
	}{
		{"1.0.0-alpha1", "1.0.0", false},
		{"1.0.0-beta2", "1.0.0", false},
		{"1.0.0-rc3", "1.0.0", false},
		{"1.0.0", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, _ := semver.NewVersion(tt.version)
			result, err := vm.ReleaseStable(v)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result.String() != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, result.String())
				}
			}
		})
	}
}

func TestGetLatestPreReleases(t *testing.T) {
	vm := NewVersionManager()

	tags := []string{
		"v1.0.0",
		"v1.0.1-alpha1",
		"v1.0.1-alpha2",
		"v1.0.1-alpha3",
		"v1.0.1-beta1",
		"v1.0.1-beta2",
		"v1.0.1-rc1",
		"v1.0.2-alpha1",
	}

	versions, _ := vm.ParseTags(tags)
	baseVersion, _ := semver.NewVersion("1.0.1")

	result := vm.GetLatestPreReleases(versions, baseVersion)

	if result.Alpha == nil || result.Alpha.String() != "1.0.1-alpha3" {
		t.Errorf("expected alpha 1.0.1-alpha3, got %v", result.Alpha)
	}
	if result.Beta == nil || result.Beta.String() != "1.0.1-beta2" {
		t.Errorf("expected beta 1.0.1-beta2, got %v", result.Beta)
	}
	if result.RC == nil || result.RC.String() != "1.0.1-rc1" {
		t.Errorf("expected rc 1.0.1-rc1, got %v", result.RC)
	}
}

func TestGetLatestPreReleasesEmpty(t *testing.T) {
	vm := NewVersionManager()

	tags := []string{"v1.0.0", "v2.0.0"}
	versions, _ := vm.ParseTags(tags)
	baseVersion, _ := semver.NewVersion("1.0.1")

	result := vm.GetLatestPreReleases(versions, baseVersion)

	if result.Alpha != nil {
		t.Errorf("expected nil alpha, got %v", result.Alpha)
	}
	if result.Beta != nil {
		t.Errorf("expected nil beta, got %v", result.Beta)
	}
	if result.RC != nil {
		t.Errorf("expected nil rc, got %v", result.RC)
	}
}

func TestFindNextPreReleaseNumber(t *testing.T) {
	vm := NewVersionManager()

	tags := []string{
		"v1.0.0",
		"v1.0.1-alpha1",
		"v1.0.1-alpha2",
		"v1.0.1-alpha3",
		"v1.0.1-beta1",
	}

	versions, _ := vm.ParseTags(tags)
	baseVersion, _ := semver.NewVersion("1.0.1")

	tests := []struct {
		preType  PreReleaseType
		expected int
	}{
		{PreReleaseAlpha, 4},
		{PreReleaseBeta, 2},
		{PreReleaseRC, 1},
	}

	for _, tt := range tests {
		t.Run(string(tt.preType), func(t *testing.T) {
			result := vm.FindNextPreReleaseNumber(versions, baseVersion, tt.preType)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetLatestStableVersion(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{
			name:     "mixed stable and pre-release",
			tags:     []string{"v1.0.0", "v1.0.1-alpha1", "v1.1.0", "v1.1.1-beta1"},
			expected: "1.1.0",
		},
		{
			name:     "only pre-releases",
			tags:     []string{"v1.0.0-alpha1", "v1.0.0-beta1"},
			expected: "0.0.0",
		},
		{
			name:     "only stable",
			tags:     []string{"v1.0.0", "v2.0.0"},
			expected: "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions, _ := vm.ParseTags(tt.tags)
			result := vm.GetLatestStableVersion(versions)
			if result.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestGetAllPreReleasesForBase(t *testing.T) {
	vm := NewVersionManager()

	tags := []string{
		"v1.0.0",
		"v1.0.1-alpha1",
		"v1.0.1-alpha2",
		"v1.0.1-beta1",
		"v1.0.2-alpha1",
	}

	versions, _ := vm.ParseTags(tags)
	baseVersion, _ := semver.NewVersion("1.0.1")

	result := vm.GetAllPreReleasesForBase(versions, baseVersion)

	if len(result) != 3 {
		t.Errorf("expected 3 pre-releases, got %d", len(result))
	}

	// Should be sorted
	expected := []string{"1.0.1-alpha1", "1.0.1-alpha2", "1.0.1-beta1"}
	for i, v := range result {
		if v.String() != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, v.String())
		}
	}
}

func TestFormatWithTemplate(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		name     string
		version  string
		template string
		expected string
	}{
		{
			name:     "default template",
			version:  "1.2.3",
			template: "v{major}.{minor}.{patch}",
			expected: "v1.2.3",
		},
		{
			name:     "custom prefix",
			version:  "2.0.0",
			template: "release-{major}.{minor}.{patch}",
			expected: "release-2.0.0",
		},
		{
			name:     "no prefix",
			version:  "1.0.0",
			template: "{major}.{minor}.{patch}",
			expected: "1.0.0",
		},
		{
			name:     "major only",
			version:  "3.5.7",
			template: "v{major}",
			expected: "v3",
		},
		{
			name:     "complex template",
			version:  "1.2.3",
			template: "version-{major}-{minor}-{patch}-release",
			expected: "version-1-2-3-release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := semver.NewVersion(tt.version)
			result := vm.FormatWithTemplate(v, tt.template)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatPreReleaseWithTemplate(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		name     string
		version  string
		template string
		preType  string
		preNum   int
		expected string
	}{
		{
			name:     "default template",
			version:  "1.2.3",
			template: "v{major}.{minor}.{patch}-{preReleaseType}{preReleaseNum}",
			preType:  "alpha",
			preNum:   1,
			expected: "v1.2.3-alpha1",
		},
		{
			name:     "beta with custom number",
			version:  "2.0.0",
			template: "v{major}.{minor}.{patch}-{preReleaseType}{preReleaseNum}",
			preType:  "beta",
			preNum:   5,
			expected: "v2.0.0-beta5",
		},
		{
			name:     "rc version",
			version:  "1.0.0",
			template: "v{major}.{minor}.{patch}-{preReleaseType}{preReleaseNum}",
			preType:  "rc",
			preNum:   3,
			expected: "v1.0.0-rc3",
		},
		{
			name:     "custom template with hyphen separator",
			version:  "1.2.3",
			template: "v{major}.{minor}.{patch}-{preReleaseType}-{preReleaseNum}",
			preType:  "alpha",
			preNum:   1,
			expected: "v1.2.3-alpha-1",
		},
		{
			name:     "custom pre-release type",
			version:  "1.0.0",
			template: "v{major}.{minor}.{patch}-{preReleaseType}{preReleaseNum}",
			preType:  "nightly",
			preNum:   20240101,
			expected: "v1.0.0-nightly20240101",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := semver.NewVersion(tt.version)
			result := vm.FormatPreReleaseWithTemplate(v, tt.template, tt.preType, tt.preNum)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatVersionWithPrefix(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		name     string
		version  string
		prefix   string
		expected string
	}{
		{
			name:     "v prefix",
			version:  "1.2.3",
			prefix:   "v",
			expected: "v1.2.3",
		},
		{
			name:     "release prefix",
			version:  "2.0.0",
			prefix:   "release-",
			expected: "release-2.0.0",
		},
		{
			name:     "empty prefix",
			version:  "1.0.0",
			prefix:   "",
			expected: "1.0.0",
		},
		{
			name:     "version prefix",
			version:  "3.1.4",
			prefix:   "version-",
			expected: "version-3.1.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := semver.NewVersion(tt.version)
			result := vm.FormatVersionWithPrefix(v, tt.prefix)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetPreReleaseInfoWithTypes(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		name         string
		version      string
		validTypes   []string
		expectedType PreReleaseType
		expectedNum  int
		isNil        bool
	}{
		{
			name:         "alpha in default types",
			version:      "1.0.0-alpha1",
			validTypes:   []string{"alpha", "beta", "rc"},
			expectedType: PreReleaseAlpha,
			expectedNum:  1,
			isNil:        false,
		},
		{
			name:         "custom type nightly",
			version:      "1.0.0-nightly5",
			validTypes:   []string{"nightly", "preview"},
			expectedType: "nightly",
			expectedNum:  5,
			isNil:        false,
		},
		{
			name:         "type not in valid list",
			version:      "1.0.0-alpha1",
			validTypes:   []string{"beta", "rc"},
			expectedType: "",
			expectedNum:  0,
			isNil:        true,
		},
		{
			name:         "stable version",
			version:      "1.0.0",
			validTypes:   []string{"alpha", "beta", "rc"},
			expectedType: "",
			expectedNum:  0,
			isNil:        true,
		},
		{
			name:         "invalid pre-release format",
			version:      "1.0.0-invalid",
			validTypes:   []string{"alpha", "beta", "rc"},
			expectedType: "",
			expectedNum:  0,
			isNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := semver.NewVersion(tt.version)
			info := vm.GetPreReleaseInfoWithTypes(v, tt.validTypes)
			if tt.isNil {
				if info != nil {
					t.Errorf("expected nil, got %+v", info)
				}
			} else {
				if info == nil {
					t.Fatal("expected info, got nil")
				}
				if info.Type != tt.expectedType {
					t.Errorf("expected type %s, got %s", tt.expectedType, info.Type)
				}
				if info.Number != tt.expectedNum {
					t.Errorf("expected number %d, got %d", tt.expectedNum, info.Number)
				}
			}
		})
	}
}

func TestSetPreReleaseString(t *testing.T) {
	vm := NewVersionManager()
	v, _ := semver.NewVersion("1.2.3")

	tests := []struct {
		name     string
		preType  string
		num      int
		expected string
	}{
		{"alpha", "alpha", 1, "1.2.3-alpha1"},
		{"beta", "beta", 2, "1.2.3-beta2"},
		{"rc", "rc", 3, "1.2.3-rc3"},
		{"custom nightly", "nightly", 10, "1.2.3-nightly10"},
		{"custom preview", "preview", 1, "1.2.3-preview1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.SetPreReleaseString(v, tt.preType, tt.num)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestNextPreReleaseStageWithTypes(t *testing.T) {
	vm := NewVersionManager()

	tests := []struct {
		name     string
		version  string
		types    []string
		expected string
		hasError bool
	}{
		{
			name:     "alpha to beta",
			version:  "1.0.0-alpha3",
			types:    []string{"alpha", "beta", "rc"},
			expected: "1.0.0-beta1",
			hasError: false,
		},
		{
			name:     "beta to rc",
			version:  "1.0.0-beta2",
			types:    []string{"alpha", "beta", "rc"},
			expected: "1.0.0-rc1",
			hasError: false,
		},
		{
			name:     "rc is final - error",
			version:  "1.0.0-rc1",
			types:    []string{"alpha", "beta", "rc"},
			expected: "",
			hasError: true,
		},
		{
			name:     "custom types: dev to staging",
			version:  "1.0.0-dev5",
			types:    []string{"dev", "staging", "prod"},
			expected: "1.0.0-staging1",
			hasError: false,
		},
		{
			name:     "unknown type - error",
			version:  "1.0.0-unknown1",
			types:    []string{"alpha", "beta", "rc"},
			expected: "",
			hasError: true,
		},
		{
			name:     "stable version - error",
			version:  "1.0.0",
			types:    []string{"alpha", "beta", "rc"},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := semver.NewVersion(tt.version)
			result, err := vm.NextPreReleaseStageWithTypes(v, tt.types)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result.String() != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, result.String())
				}
			}
		})
	}
}

func TestFindNextPreReleaseNumberForType(t *testing.T) {
	vm := NewVersionManager()

	tags := []string{
		"v1.0.0",
		"v1.0.1-alpha1",
		"v1.0.1-alpha2",
		"v1.0.1-alpha3",
		"v1.0.1-beta1",
		"v1.0.1-nightly5",
	}

	versions, _ := vm.ParseTags(tags)
	baseVersion, _ := semver.NewVersion("1.0.1")

	tests := []struct {
		preType  string
		expected int
	}{
		{"alpha", 4},
		{"beta", 2},
		{"rc", 1},
		{"nightly", 6},
		{"preview", 1}, // no existing preview versions
	}

	for _, tt := range tests {
		t.Run(tt.preType, func(t *testing.T) {
			result := vm.FindNextPreReleaseNumberForType(versions, baseVersion, tt.preType)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetLatestPreReleasesForTypes(t *testing.T) {
	vm := NewVersionManager()

	tags := []string{
		"v1.0.0",
		"v1.0.1-alpha1",
		"v1.0.1-alpha2",
		"v1.0.1-alpha3",
		"v1.0.1-beta1",
		"v1.0.1-beta2",
		"v1.0.1-rc1",
		"v1.0.1-nightly5",
		"v1.0.2-alpha1",
	}

	versions, _ := vm.ParseTags(tags)
	baseVersion, _ := semver.NewVersion("1.0.1")

	t.Run("default types", func(t *testing.T) {
		types := []string{"alpha", "beta", "rc"}
		result := vm.GetLatestPreReleasesForTypes(versions, baseVersion, types)

		if result["alpha"] == nil || result["alpha"].String() != "1.0.1-alpha3" {
			t.Errorf("expected alpha 1.0.1-alpha3, got %v", result["alpha"])
		}
		if result["beta"] == nil || result["beta"].String() != "1.0.1-beta2" {
			t.Errorf("expected beta 1.0.1-beta2, got %v", result["beta"])
		}
		if result["rc"] == nil || result["rc"].String() != "1.0.1-rc1" {
			t.Errorf("expected rc 1.0.1-rc1, got %v", result["rc"])
		}
	})

	t.Run("custom types including nightly", func(t *testing.T) {
		types := []string{"alpha", "nightly"}
		result := vm.GetLatestPreReleasesForTypes(versions, baseVersion, types)

		if result["alpha"] == nil || result["alpha"].String() != "1.0.1-alpha3" {
			t.Errorf("expected alpha 1.0.1-alpha3, got %v", result["alpha"])
		}
		if result["nightly"] == nil || result["nightly"].String() != "1.0.1-nightly5" {
			t.Errorf("expected nightly 1.0.1-nightly5, got %v", result["nightly"])
		}
	})

	t.Run("no matching types", func(t *testing.T) {
		types := []string{"preview", "snapshot"}
		result := vm.GetLatestPreReleasesForTypes(versions, baseVersion, types)

		if result["preview"] != nil {
			t.Errorf("expected nil preview, got %v", result["preview"])
		}
		if result["snapshot"] != nil {
			t.Errorf("expected nil snapshot, got %v", result["snapshot"])
		}
	})

	t.Run("empty types", func(t *testing.T) {
		types := []string{}
		result := vm.GetLatestPreReleasesForTypes(versions, baseVersion, types)

		if len(result) != 0 {
			t.Errorf("expected empty map, got %d entries", len(result))
		}
	})
}
