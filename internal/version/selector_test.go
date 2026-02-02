package version

import (
	"testing"

	sv "github.com/Masterminds/semver/v3"
)

func TestNewSelector(t *testing.T) {
	s := NewSelector()
	if s == nil {
		t.Error("NewSelector() returned nil")
	}
	if s.vm == nil {
		t.Error("NewSelector() vm is nil")
	}
}

func TestSelector_GetPreReleaseOptions(t *testing.T) {
	s := NewSelector()

	tests := []struct {
		name        string
		current     string
		versions    []string
		wantLen     int
		wantActions []PreReleaseAction
		wantErr     bool
	}{
		{
			name:        "alpha version - all three options",
			current:     "1.0.0-alpha1",
			versions:    []string{"1.0.0-alpha1"},
			wantLen:     3,
			wantActions: []PreReleaseAction{PreReleaseActionBump, PreReleaseActionAdvance, PreReleaseActionStable},
			wantErr:     false,
		},
		{
			name:        "beta version - all three options",
			current:     "1.0.0-beta2",
			versions:    []string{"1.0.0-beta2"},
			wantLen:     3,
			wantActions: []PreReleaseAction{PreReleaseActionBump, PreReleaseActionAdvance, PreReleaseActionStable},
			wantErr:     false,
		},
		{
			name:        "rc version - no advance option",
			current:     "1.0.0-rc1",
			versions:    []string{"1.0.0-rc1"},
			wantLen:     2,
			wantActions: []PreReleaseAction{PreReleaseActionBump, PreReleaseActionStable},
			wantErr:     false,
		},
		{
			name:        "stable version - error",
			current:     "1.0.0",
			versions:    []string{"1.0.0"},
			wantLen:     0,
			wantActions: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, _ := sv.NewVersion(tt.current)
			var versions []*sv.Version
			for _, v := range tt.versions {
				ver, _ := sv.NewVersion(v)
				versions = append(versions, ver)
			}

			options, err := s.GetPreReleaseOptions(current, versions)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetPreReleaseOptions() error = nil, wantErr true")
				}
				return
			}

			if err != nil {
				t.Errorf("GetPreReleaseOptions() unexpected error = %v", err)
				return
			}

			if len(options) != tt.wantLen {
				t.Errorf("GetPreReleaseOptions() returned %d options, want %d", len(options), tt.wantLen)
			}

			for i, wantAction := range tt.wantActions {
				if i >= len(options) {
					break
				}
				if options[i].Action != wantAction {
					t.Errorf("GetPreReleaseOptions() option[%d].Action = %v, want %v", i, options[i].Action, wantAction)
				}
			}
		})
	}
}

func TestSelector_GetPreReleaseOptions_Content(t *testing.T) {
	s := NewSelector()

	t.Run("alpha version generates correct versions", func(t *testing.T) {
		current, _ := sv.NewVersion("1.0.0-alpha1")
		versions := []*sv.Version{current}

		options, err := s.GetPreReleaseOptions(current, versions)
		if err != nil {
			t.Fatalf("GetPreReleaseOptions() error = %v", err)
		}

		// Check bump option
		if options[0].NewVersion != "v1.0.0-alpha2" {
			t.Errorf("bump option version = %s, want v1.0.0-alpha2", options[0].NewVersion)
		}
		if options[0].Desc != "increment alpha number" {
			t.Errorf("bump option desc = %s, want 'increment alpha number'", options[0].Desc)
		}

		// Check advance option
		if options[1].NewVersion != "v1.0.0-beta1" {
			t.Errorf("advance option version = %s, want v1.0.0-beta1", options[1].NewVersion)
		}
		if options[1].Desc != "alpha → beta" {
			t.Errorf("advance option desc = %s, want 'alpha → beta'", options[1].Desc)
		}

		// Check stable option
		if options[2].NewVersion != "v1.0.0" {
			t.Errorf("stable option version = %s, want v1.0.0", options[2].NewVersion)
		}
	})
}

func TestSelector_GetStableBumpOptions(t *testing.T) {
	s := NewSelector()

	tests := []struct {
		name         string
		current      string
		wantTypes    []string
		wantVersions []string
	}{
		{
			name:         "stable version options",
			current:      "1.2.3",
			wantTypes:    []string{"patch", "minor", "major"},
			wantVersions: []string{"v1.2.4", "v1.3.0", "v2.0.0"},
		},
		{
			name:         "zero version options",
			current:      "0.0.0",
			wantTypes:    []string{"patch", "minor", "major"},
			wantVersions: []string{"v0.0.1", "v0.1.0", "v1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, _ := sv.NewVersion(tt.current)
			options := s.GetStableBumpOptions(current)

			if len(options) != 3 {
				t.Errorf("GetStableBumpOptions() returned %d options, want 3", len(options))
			}

			for i, opt := range options {
				if opt.Type != tt.wantTypes[i] {
					t.Errorf("option[%d].Type = %s, want %s", i, opt.Type, tt.wantTypes[i])
				}
				if opt.NewVersion != tt.wantVersions[i] {
					t.Errorf("option[%d].NewVersion = %s, want %s", i, opt.NewVersion, tt.wantVersions[i])
				}
			}
		})
	}
}

func TestSelector_GetStableBumpOptions_Desc(t *testing.T) {
	s := NewSelector()
	current, _ := sv.NewVersion("1.0.0")
	options := s.GetStableBumpOptions(current)

	wantDescs := []string{"补丁更新", "小版本更新", "大版本更新"}
	for i, opt := range options {
		if opt.Desc != wantDescs[i] {
			t.Errorf("option[%d].Desc = %s, want %s", i, opt.Desc, wantDescs[i])
		}
	}
}

func TestSelector_CalculateBumpedVersion(t *testing.T) {
	s := NewSelector()

	tests := []struct {
		name     string
		current  string
		bumpType string
		want     string
		wantErr  bool
	}{
		{"patch bump", "1.0.0", "patch", "1.0.1", false},
		{"minor bump", "1.0.0", "minor", "1.1.0", false},
		{"major bump", "1.0.0", "major", "2.0.0", false},
		{"uppercase PATCH", "1.0.0", "PATCH", "1.0.1", false},
		{"uppercase MINOR", "1.0.0", "MINOR", "1.1.0", false},
		{"uppercase MAJOR", "1.0.0", "MAJOR", "2.0.0", false},
		{"invalid type", "1.0.0", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, _ := sv.NewVersion(tt.current)
			got, err := s.CalculateBumpedVersion(current, tt.bumpType)

			if tt.wantErr {
				if err == nil {
					t.Error("CalculateBumpedVersion() error = nil, wantErr true")
				}
				return
			}

			if err != nil {
				t.Errorf("CalculateBumpedVersion() unexpected error = %v", err)
				return
			}

			if got.String() != tt.want {
				t.Errorf("CalculateBumpedVersion() = %s, want %s", got.String(), tt.want)
			}
		})
	}
}

func TestSelector_GetPreReleaseTypeOptions(t *testing.T) {
	s := NewSelector()

	tests := []struct {
		name           string
		baseVersion    string
		versions       []string
		wantTypes      []string
		wantVersions   []string
		wantLatestInfo []string
	}{
		{
			name:           "new pre-release from scratch",
			baseVersion:    "1.0.1",
			versions:       []string{"1.0.0"},
			wantTypes:      []string{"alpha", "beta", "rc"},
			wantVersions:   []string{"v1.0.1-alpha1", "v1.0.1-beta1", "v1.0.1-rc1"},
			wantLatestInfo: []string{"", "", ""},
		},
		{
			name:           "with existing pre-releases",
			baseVersion:    "1.0.0",
			versions:       []string{"1.0.0-alpha2", "1.0.0-beta1"},
			wantTypes:      []string{"alpha", "beta", "rc"},
			wantVersions:   []string{"v1.0.0-alpha3", "v1.0.0-beta2", "v1.0.0-rc1"},
			wantLatestInfo: []string{"v1.0.0-alpha2", "v1.0.0-beta1", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, _ := sv.NewVersion(tt.baseVersion)
			var versions []*sv.Version
			for _, v := range tt.versions {
				ver, _ := sv.NewVersion(v)
				versions = append(versions, ver)
			}

			options := s.GetPreReleaseTypeOptions(base, versions)

			if len(options) != 3 {
				t.Errorf("GetPreReleaseTypeOptions() returned %d options, want 3", len(options))
			}

			for i, opt := range options {
				if opt.Type != tt.wantTypes[i] {
					t.Errorf("option[%d].Type = %v, want %v", i, opt.Type, tt.wantTypes[i])
				}
				if opt.NewVersion != tt.wantVersions[i] {
					t.Errorf("option[%d].NewVersion = %s, want %s", i, opt.NewVersion, tt.wantVersions[i])
				}
				if opt.LatestInfo != tt.wantLatestInfo[i] {
					t.Errorf("option[%d].LatestInfo = %s, want %s", i, opt.LatestInfo, tt.wantLatestInfo[i])
				}
			}
		})
	}
}

func TestSelector_GetPreReleaseTypeOptions_FilterByBaseVersion(t *testing.T) {
	s := NewSelector()

	// Only pre-releases for 1.0.0 should be considered, not 1.0.1
	base, _ := sv.NewVersion("1.0.0")
	versions := []*sv.Version{
		mustParse("1.0.0-alpha5"),
		mustParse("1.0.0-alpha3"),
		mustParse("1.0.1-alpha1"), // should be ignored
	}

	options := s.GetPreReleaseTypeOptions(base, versions)

	// Alpha should suggest alpha6 (next after alpha5)
	if options[0].NewVersion != "v1.0.0-alpha6" {
		t.Errorf("alpha option = %s, want v1.0.0-alpha6", options[0].NewVersion)
	}
	if options[0].LatestInfo != "v1.0.0-alpha5" {
		t.Errorf("alpha latestInfo = %s, want v1.0.0-alpha5", options[0].LatestInfo)
	}
}

func TestSelector_CalculateNonInteractiveVersion(t *testing.T) {
	s := NewSelector()

	tests := []struct {
		name           string
		current        string
		versions       []string
		preReleaseType string
		preReleaseNum  int
		want           string
		wantErr        bool
	}{
		{
			name:           "stable to alpha",
			current:        "1.0.0",
			versions:       []string{"1.0.0"},
			preReleaseType: "alpha",
			preReleaseNum:  1,
			want:           "v1.0.1-alpha1",
			wantErr:        false,
		},
		{
			name:           "stable to beta with custom number",
			current:        "1.0.0",
			versions:       []string{"1.0.0"},
			preReleaseType: "beta",
			preReleaseNum:  5,
			want:           "v1.0.1-beta5",
			wantErr:        false,
		},
		{
			name:           "pre-release to alpha",
			current:        "1.0.0-beta1",
			versions:       []string{"1.0.0-beta1"},
			preReleaseType: "alpha",
			preReleaseNum:  1,
			want:           "v1.0.0-alpha1",
			wantErr:        false,
		},
		{
			name:           "zero number defaults to 1",
			current:        "1.0.0",
			versions:       []string{"1.0.0"},
			preReleaseType: "rc",
			preReleaseNum:  0,
			want:           "v1.0.1-rc1",
			wantErr:        false,
		},
		{
			name:           "negative number defaults to 1",
			current:        "1.0.0",
			versions:       []string{"1.0.0"},
			preReleaseType: "rc",
			preReleaseNum:  -5,
			want:           "v1.0.1-rc1",
			wantErr:        false,
		},
		{
			name:           "invalid pre-release type",
			current:        "1.0.0",
			versions:       []string{"1.0.0"},
			preReleaseType: "invalid",
			preReleaseNum:  1,
			want:           "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, _ := sv.NewVersion(tt.current)
			var versions []*sv.Version
			for _, v := range tt.versions {
				ver, _ := sv.NewVersion(v)
				versions = append(versions, ver)
			}

			got, err := s.CalculateNonInteractiveVersion(current, versions, tt.preReleaseType, tt.preReleaseNum)

			if tt.wantErr {
				if err == nil {
					t.Error("CalculateNonInteractiveVersion() error = nil, wantErr true")
				}
				return
			}

			if err != nil {
				t.Errorf("CalculateNonInteractiveVersion() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("CalculateNonInteractiveVersion() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestSelector_GetVersionManager(t *testing.T) {
	s := NewSelector()
	vm := s.GetVersionManager()

	if vm == nil {
		t.Error("GetVersionManager() returned nil")
	}

	// Test that the returned VM works
	v, _ := sv.NewVersion("1.0.0")
	formatted := vm.FormatVersion(v)
	if formatted != "v1.0.0" {
		t.Errorf("VersionManager.FormatVersion() = %s, want v1.0.0", formatted)
	}
}

func TestSelector_CalculateNonInteractiveVersion_PreReleaseVariations(t *testing.T) {
	s := NewSelector()

	tests := []struct {
		name           string
		current        string
		preReleaseType string
		preReleaseNum  int
		want           string
	}{
		{
			name:           "pre-release to beta",
			current:        "1.0.0-alpha1",
			preReleaseType: "beta",
			preReleaseNum:  1,
			want:           "v1.0.0-beta1",
		},
		{
			name:           "pre-release to rc",
			current:        "1.0.0-beta2",
			preReleaseType: "rc",
			preReleaseNum:  3,
			want:           "v1.0.0-rc3",
		},
		{
			name:           "same pre-release type bump",
			current:        "1.0.0-alpha5",
			preReleaseType: "alpha",
			preReleaseNum:  10,
			want:           "v1.0.0-alpha10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, _ := sv.NewVersion(tt.current)
			versions := []*sv.Version{current}

			got, err := s.CalculateNonInteractiveVersion(current, versions, tt.preReleaseType, tt.preReleaseNum)
			if err != nil {
				t.Errorf("CalculateNonInteractiveVersion() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("CalculateNonInteractiveVersion() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestSelector_GetPreReleaseTypeOptions_WithAllRC(t *testing.T) {
	s := NewSelector()

	base, _ := sv.NewVersion("2.0.0")
	versions := []*sv.Version{
		mustParse("2.0.0-rc1"),
		mustParse("2.0.0-rc3"),
	}

	options := s.GetPreReleaseTypeOptions(base, versions)

	// RC should show rc4 as next
	rcOption := options[2]
	if rcOption.Type != "rc" {
		t.Errorf("expected RC type, got %v", rcOption.Type)
	}
	if rcOption.NewVersion != "v2.0.0-rc4" {
		t.Errorf("RC next version = %s, want v2.0.0-rc4", rcOption.NewVersion)
	}
	if rcOption.LatestInfo != "v2.0.0-rc3" {
		t.Errorf("RC latest info = %s, want v2.0.0-rc3", rcOption.LatestInfo)
	}
}

func TestSelector_GetPreReleaseOptions_EdgeCases(t *testing.T) {
	s := NewSelector()

	t.Run("high number pre-release", func(t *testing.T) {
		current, _ := sv.NewVersion("1.0.0-alpha999")
		versions := []*sv.Version{current}

		options, err := s.GetPreReleaseOptions(current, versions)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check bump suggests alpha1000
		if options[0].NewVersion != "v1.0.0-alpha1000" {
			t.Errorf("bump version = %s, want v1.0.0-alpha1000", options[0].NewVersion)
		}
	})

	t.Run("beta to rc advance", func(t *testing.T) {
		current, _ := sv.NewVersion("1.0.0-beta5")
		versions := []*sv.Version{current}

		options, err := s.GetPreReleaseOptions(current, versions)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have 3 options: bump, advance, stable
		if len(options) != 3 {
			t.Errorf("expected 3 options, got %d", len(options))
		}

		// Check advance option
		advanceFound := false
		for _, opt := range options {
			if opt.Action == PreReleaseActionAdvance {
				advanceFound = true
				if opt.NewVersion != "v1.0.0-rc1" {
					t.Errorf("advance version = %s, want v1.0.0-rc1", opt.NewVersion)
				}
				if opt.Desc != "beta → rc" {
					t.Errorf("advance desc = %s, want 'beta → rc'", opt.Desc)
				}
			}
		}
		if !advanceFound {
			t.Error("advance option not found")
		}
	})
}

func TestSelector_GetStableBumpOptions_PreReleaseBase(t *testing.T) {
	s := NewSelector()

	// From pre-release, bump options work on the base version (pre-release stripped)
	// Note: Masterminds/semver IncPatch/IncMinor/IncMajor on pre-release strips the pre-release
	// but doesn't increment the version number - this is the library's behavior
	current, _ := sv.NewVersion("1.5.3-alpha1")
	options := s.GetStableBumpOptions(current)

	// For pre-release 1.5.3-alpha1:
	// - IncPatch returns 1.5.3 (strips pre-release, no increment)
	// - IncMinor returns 1.6.0
	// - IncMajor returns 2.0.0
	wantVersions := []string{"v1.5.3", "v1.6.0", "v2.0.0"}
	for i, opt := range options {
		if opt.NewVersion != wantVersions[i] {
			t.Errorf("option[%d].NewVersion = %s, want %s", i, opt.NewVersion, wantVersions[i])
		}
	}
}

// Helper function
func mustParse(s string) *sv.Version {
	v, _ := sv.NewVersion(s)
	return v
}
