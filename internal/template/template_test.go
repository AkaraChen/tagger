package template

import (
	"testing"
)

func TestNewTagTemplate_Defaults(t *testing.T) {
	tmpl, err := NewTagTemplate("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tmpl.ReleaseTemplate() != DefaultReleaseTemplate {
		t.Errorf("expected release template %q, got %q", DefaultReleaseTemplate, tmpl.ReleaseTemplate())
	}
	if tmpl.PreReleaseTemplate() != DefaultPreReleaseTemplate {
		t.Errorf("expected pre-release template %q, got %q", DefaultPreReleaseTemplate, tmpl.PreReleaseTemplate())
	}
}

func TestNewTagTemplate_InvalidRelease(t *testing.T) {
	_, err := NewTagTemplate("v{major}.{minor}", "") // missing {patch}
	if err == nil {
		t.Error("expected error for missing patch placeholder")
	}
}

func TestNewTagTemplate_InvalidPreRelease(t *testing.T) {
	_, err := NewTagTemplate("", "v{major}.{minor}.{patch}-{preType}") // missing {preNum}
	if err == nil {
		t.Error("expected error for missing preNum placeholder")
	}
}

func TestTagTemplate_Parse_DefaultRelease(t *testing.T) {
	tmpl := NewDefaultTagTemplate()

	tests := []struct {
		tag     string
		major   uint64
		minor   uint64
		patch   uint64
		wantErr bool
	}{
		{"v1.2.3", 1, 2, 3, false},
		{"v0.0.0", 0, 0, 0, false},
		{"v10.20.30", 10, 20, 30, false},
		{"v999.999.999", 999, 999, 999, false},
		{"1.2.3", 0, 0, 0, true},   // missing v prefix
		{"v1.2", 0, 0, 0, true},    // missing patch
		{"vx.y.z", 0, 0, 0, true},  // non-numeric
		{"", 0, 0, 0, true},        // empty
		{"v1.2.3.4", 0, 0, 0, true}, // extra component
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			parsed, err := tmpl.Parse(tt.tag)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for tag %q", tt.tag)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if parsed.Major != tt.major || parsed.Minor != tt.minor || parsed.Patch != tt.patch {
				t.Errorf("got %d.%d.%d, want %d.%d.%d", parsed.Major, parsed.Minor, parsed.Patch, tt.major, tt.minor, tt.patch)
			}
			if parsed.IsPreRelease() {
				t.Error("expected non-pre-release")
			}
		})
	}
}

func TestTagTemplate_Parse_DefaultPreRelease(t *testing.T) {
	tmpl := NewDefaultTagTemplate()

	tests := []struct {
		tag     string
		major   uint64
		minor   uint64
		patch   uint64
		preType string
		preNum  int
		wantErr bool
	}{
		{"v1.0.0-alpha-1", 1, 0, 0, "alpha", 1, false},
		{"v2.3.4-beta-5", 2, 3, 4, "beta", 5, false},
		{"v1.0.0-rc-1", 1, 0, 0, "rc", 1, false},
		{"v0.1.0-dev-99", 0, 1, 0, "dev", 99, false},
		{"v1.0.0-alpha", 0, 0, 0, "", 0, true},     // missing number
		{"v1.0.0-1", 0, 0, 0, "", 0, true},         // missing type
		{"v1.0.0-alpha-", 0, 0, 0, "", 0, true},    // invalid number
		{"v1.0.0-alpha-abc", 0, 0, 0, "", 0, true}, // non-numeric preNum
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			parsed, err := tmpl.Parse(tt.tag)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for tag %q", tt.tag)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if parsed.Major != tt.major || parsed.Minor != tt.minor || parsed.Patch != tt.patch {
				t.Errorf("got %d.%d.%d, want %d.%d.%d", parsed.Major, parsed.Minor, parsed.Patch, tt.major, tt.minor, tt.patch)
			}
			if parsed.PreType != tt.preType {
				t.Errorf("got preType %q, want %q", parsed.PreType, tt.preType)
			}
			if parsed.PreNum != tt.preNum {
				t.Errorf("got preNum %d, want %d", parsed.PreNum, tt.preNum)
			}
			if !parsed.IsPreRelease() {
				t.Error("expected pre-release")
			}
		})
	}
}

func TestTagTemplate_Format_Default(t *testing.T) {
	tmpl := NewDefaultTagTemplate()

	tests := []struct {
		major uint64
		minor uint64
		patch uint64
		want  string
	}{
		{1, 2, 3, "v1.2.3"},
		{0, 0, 0, "v0.0.0"},
		{10, 20, 30, "v10.20.30"},
	}

	for _, tt := range tests {
		got := tmpl.Format(tt.major, tt.minor, tt.patch)
		if got != tt.want {
			t.Errorf("Format(%d, %d, %d) = %q, want %q", tt.major, tt.minor, tt.patch, got, tt.want)
		}
	}
}

func TestTagTemplate_FormatPreRelease_Default(t *testing.T) {
	tmpl := NewDefaultTagTemplate()

	tests := []struct {
		major   uint64
		minor   uint64
		patch   uint64
		preType string
		preNum  int
		want    string
	}{
		{1, 0, 0, "alpha", 1, "v1.0.0-alpha-1"},
		{2, 3, 4, "beta", 5, "v2.3.4-beta-5"},
		{1, 0, 0, "rc", 10, "v1.0.0-rc-10"},
	}

	for _, tt := range tests {
		got := tmpl.FormatPreRelease(tt.major, tt.minor, tt.patch, tt.preType, tt.preNum)
		if got != tt.want {
			t.Errorf("FormatPreRelease(%d, %d, %d, %q, %d) = %q, want %q",
				tt.major, tt.minor, tt.patch, tt.preType, tt.preNum, got, tt.want)
		}
	}
}

func TestTagTemplate_CustomTemplate_NoPrefix(t *testing.T) {
	tmpl, err := NewTagTemplate(
		"{major}.{minor}.{patch}",
		"{major}.{minor}.{patch}-{preType}.{preNum}",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test parsing
	parsed, err := tmpl.Parse("1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Major != 1 || parsed.Minor != 2 || parsed.Patch != 3 {
		t.Errorf("got %d.%d.%d, want 1.2.3", parsed.Major, parsed.Minor, parsed.Patch)
	}

	// Test pre-release parsing
	parsed, err = tmpl.Parse("1.0.0-alpha.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.PreType != "alpha" || parsed.PreNum != 1 {
		t.Errorf("got %s.%d, want alpha.1", parsed.PreType, parsed.PreNum)
	}

	// Test formatting
	got := tmpl.Format(1, 2, 3)
	if got != "1.2.3" {
		t.Errorf("Format = %q, want %q", got, "1.2.3")
	}

	got = tmpl.FormatPreRelease(1, 0, 0, "alpha", 1)
	if got != "1.0.0-alpha.1" {
		t.Errorf("FormatPreRelease = %q, want %q", got, "1.0.0-alpha.1")
	}
}

func TestTagTemplate_CustomTemplate_WithPrefix(t *testing.T) {
	tmpl, err := NewTagTemplate(
		"release/v{major}.{minor}.{patch}",
		"release/v{major}.{minor}.{patch}-{preType}-{preNum}",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test parsing
	parsed, err := tmpl.Parse("release/v1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Major != 1 || parsed.Minor != 2 || parsed.Patch != 3 {
		t.Errorf("got %d.%d.%d, want 1.2.3", parsed.Major, parsed.Minor, parsed.Patch)
	}

	// Test pre-release parsing
	parsed, err = tmpl.Parse("release/v1.0.0-alpha-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.PreType != "alpha" || parsed.PreNum != 1 {
		t.Errorf("got %s-%d, want alpha-1", parsed.PreType, parsed.PreNum)
	}

	// Test formatting
	got := tmpl.Format(1, 2, 3)
	if got != "release/v1.2.3" {
		t.Errorf("Format = %q, want %q", got, "release/v1.2.3")
	}

	got = tmpl.FormatPreRelease(1, 0, 0, "alpha", 1)
	if got != "release/v1.0.0-alpha-1" {
		t.Errorf("FormatPreRelease = %q, want %q", got, "release/v1.0.0-alpha-1")
	}
}

func TestTagTemplate_RoundTrip(t *testing.T) {
	templates := []struct {
		release    string
		preRelease string
	}{
		{DefaultReleaseTemplate, DefaultPreReleaseTemplate},
		{"{major}.{minor}.{patch}", "{major}.{minor}.{patch}-{preType}.{preNum}"},
		{"release/v{major}.{minor}.{patch}", "release/v{major}.{minor}.{patch}-{preType}-{preNum}"},
		{"ver-{major}-{minor}-{patch}", "ver-{major}-{minor}-{patch}-{preType}-{preNum}"},
	}

	for _, tt := range templates {
		tmpl, err := NewTagTemplate(tt.release, tt.preRelease)
		if err != nil {
			t.Fatalf("unexpected error for template %q: %v", tt.release, err)
		}

		// Round-trip test for release
		formatted := tmpl.Format(1, 2, 3)
		parsed, err := tmpl.Parse(formatted)
		if err != nil {
			t.Errorf("failed to parse formatted release tag %q: %v", formatted, err)
			continue
		}
		if parsed.Major != 1 || parsed.Minor != 2 || parsed.Patch != 3 {
			t.Errorf("round-trip failed: got %d.%d.%d, want 1.2.3", parsed.Major, parsed.Minor, parsed.Patch)
		}

		// Round-trip test for pre-release
		formatted = tmpl.FormatPreRelease(1, 0, 0, "alpha", 5)
		parsed, err = tmpl.Parse(formatted)
		if err != nil {
			t.Errorf("failed to parse formatted pre-release tag %q: %v", formatted, err)
			continue
		}
		if parsed.Major != 1 || parsed.Minor != 0 || parsed.Patch != 0 || parsed.PreType != "alpha" || parsed.PreNum != 5 {
			t.Errorf("round-trip failed: got %d.%d.%d-%s-%d, want 1.0.0-alpha-5",
				parsed.Major, parsed.Minor, parsed.Patch, parsed.PreType, parsed.PreNum)
		}
	}
}

func TestTagTemplate_CanParse(t *testing.T) {
	tmpl := NewDefaultTagTemplate()

	tests := []struct {
		tag  string
		want bool
	}{
		{"v1.2.3", true},
		{"v1.0.0-alpha-1", true},
		{"invalid", false},
		{"1.2.3", false},
		{"", false},
	}

	for _, tt := range tests {
		got := tmpl.CanParse(tt.tag)
		if got != tt.want {
			t.Errorf("CanParse(%q) = %v, want %v", tt.tag, got, tt.want)
		}
	}
}

func TestNewDefaultTagTemplate(t *testing.T) {
	tmpl := NewDefaultTagTemplate()
	if tmpl == nil {
		t.Fatal("NewDefaultTagTemplate returned nil")
	}

	// Verify it works correctly
	parsed, err := tmpl.Parse("v1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Major != 1 || parsed.Minor != 2 || parsed.Patch != 3 {
		t.Errorf("got %d.%d.%d, want 1.2.3", parsed.Major, parsed.Minor, parsed.Patch)
	}
}

func TestTagTemplate_SpecialCharsInTemplate(t *testing.T) {
	// Test that regex special characters are properly escaped
	tmpl, err := NewTagTemplate(
		"v{major}.{minor}.{patch}",
		"v{major}.{minor}.{patch}+{preType}.{preNum}",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The + should be literal, not regex quantifier
	parsed, err := tmpl.Parse("v1.0.0+alpha.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.PreType != "alpha" || parsed.PreNum != 1 {
		t.Errorf("got %s.%d, want alpha.1", parsed.PreType, parsed.PreNum)
	}

	// Should not match without +
	if tmpl.CanParse("v1.0.0-alpha.1") {
		t.Error("should not parse v1.0.0-alpha.1 with + template")
	}
}
