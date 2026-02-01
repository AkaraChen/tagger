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
			tags:     []string{"v1.0.0-alpha-1", "v1.0.0-beta-1", "v1.0.0-rc-1"},
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
			tags:     []string{"v1.0.0", "v1.0.0-alpha-1"},
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
		{"1.0.0-alpha-1", true},
		{"1.0.0-beta-2", true},
		{"1.0.0-rc-1", true},
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
		{"1.0.0-alpha-1", PreReleaseAlpha, 1, false},
		{"1.0.0-alpha-5", PreReleaseAlpha, 5, false},
		{"1.0.0-beta-2", PreReleaseBeta, 2, false},
		{"1.0.0-rc-3", PreReleaseRC, 3, false},
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
		{PreReleaseAlpha, 1, "1.2.3-alpha-1"},
		{PreReleaseBeta, 2, "1.2.3-beta-2"},
		{PreReleaseRC, 3, "1.2.3-rc-3"},
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
		{"1.0.0-alpha-1", "1.0.0-alpha-2", false},
		{"1.0.0-beta-5", "1.0.0-beta-6", false},
		{"1.0.0-rc-1", "1.0.0-rc-2", false},
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
		{"1.0.0-alpha-3", "1.0.0-beta-1", false},
		{"1.0.0-beta-2", "1.0.0-rc-1", false},
		{"1.0.0-rc-1", "", true},
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
		{"1.0.0-alpha-1", "1.0.0", false},
		{"1.0.0-beta-2", "1.0.0", false},
		{"1.0.0-rc-3", "1.0.0", false},
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
		"v1.0.1-alpha-1",
		"v1.0.1-alpha-2",
		"v1.0.1-alpha-3",
		"v1.0.1-beta-1",
		"v1.0.1-beta-2",
		"v1.0.1-rc-1",
		"v1.0.2-alpha-1",
	}

	versions, _ := vm.ParseTags(tags)
	baseVersion, _ := semver.NewVersion("1.0.1")

	result := vm.GetLatestPreReleases(versions, baseVersion)

	if result.Alpha == nil || result.Alpha.String() != "1.0.1-alpha-3" {
		t.Errorf("expected alpha 1.0.1-alpha-3, got %v", result.Alpha)
	}
	if result.Beta == nil || result.Beta.String() != "1.0.1-beta-2" {
		t.Errorf("expected beta 1.0.1-beta-2, got %v", result.Beta)
	}
	if result.RC == nil || result.RC.String() != "1.0.1-rc-1" {
		t.Errorf("expected rc 1.0.1-rc-1, got %v", result.RC)
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
		"v1.0.1-alpha-1",
		"v1.0.1-alpha-2",
		"v1.0.1-alpha-3",
		"v1.0.1-beta-1",
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
			tags:     []string{"v1.0.0", "v1.0.1-alpha-1", "v1.1.0", "v1.1.1-beta-1"},
			expected: "1.1.0",
		},
		{
			name:     "only pre-releases",
			tags:     []string{"v1.0.0-alpha-1", "v1.0.0-beta-1"},
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
		"v1.0.1-alpha-1",
		"v1.0.1-alpha-2",
		"v1.0.1-beta-1",
		"v1.0.2-alpha-1",
	}

	versions, _ := vm.ParseTags(tags)
	baseVersion, _ := semver.NewVersion("1.0.1")

	result := vm.GetAllPreReleasesForBase(versions, baseVersion)

	if len(result) != 3 {
		t.Errorf("expected 3 pre-releases, got %d", len(result))
	}

	// Should be sorted
	expected := []string{"1.0.1-alpha-1", "1.0.1-alpha-2", "1.0.1-beta-1"}
	for i, v := range result {
		if v.String() != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, v.String())
		}
	}
}
