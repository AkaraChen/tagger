package ui

import (
	"testing"

	"github.com/AkaraChen/tagger/internal/semver"
)

func TestBumpTypeOption(t *testing.T) {
	opt := BumpTypeOption{
		Type:       "patch",
		NewVersion: "v1.0.1",
		Desc:       "补丁更新",
	}

	if opt.Type != "patch" {
		t.Errorf("expected 'patch', got '%s'", opt.Type)
	}
	if opt.NewVersion != "v1.0.1" {
		t.Errorf("expected 'v1.0.1', got '%s'", opt.NewVersion)
	}
	if opt.Desc != "补丁更新" {
		t.Errorf("expected '补丁更新', got '%s'", opt.Desc)
	}
}

func TestPreReleaseTypeOption(t *testing.T) {
	opt := PreReleaseTypeOption{
		Type:       semver.PreReleaseAlpha,
		NewVersion: "v1.0.1-alpha-1",
		LatestInfo: "v1.0.1-alpha-0",
	}

	if opt.Type != semver.PreReleaseAlpha {
		t.Errorf("expected 'alpha', got '%s'", opt.Type)
	}
	if opt.NewVersion != "v1.0.1-alpha-1" {
		t.Errorf("expected 'v1.0.1-alpha-1', got '%s'", opt.NewVersion)
	}
	if opt.LatestInfo != "v1.0.1-alpha-0" {
		t.Errorf("expected 'v1.0.1-alpha-0', got '%s'", opt.LatestInfo)
	}
}

func TestPreReleaseAction(t *testing.T) {
	tests := []struct {
		action   PreReleaseAction
		expected string
	}{
		{PreReleaseActionBump, "bump"},
		{PreReleaseActionAdvance, "advance"},
		{PreReleaseActionStable, "stable"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.action) != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, string(tt.action))
			}
		})
	}
}

func TestPreReleaseActionOption(t *testing.T) {
	opt := PreReleaseActionOption{
		Action:     PreReleaseActionBump,
		NewVersion: "v1.0.1-alpha-2",
		Desc:       "increment alpha number",
	}

	if opt.Action != PreReleaseActionBump {
		t.Errorf("expected 'bump', got '%s'", opt.Action)
	}
	if opt.NewVersion != "v1.0.1-alpha-2" {
		t.Errorf("expected 'v1.0.1-alpha-2', got '%s'", opt.NewVersion)
	}
	if opt.Desc != "increment alpha number" {
		t.Errorf("expected 'increment alpha number', got '%s'", opt.Desc)
	}
}

func TestItem(t *testing.T) {
	i := item{
		title: "test title",
		desc:  "test description",
	}

	if i.Title() != "test title" {
		t.Errorf("expected 'test title', got '%s'", i.Title())
	}
	if i.Description() != "test description" {
		t.Errorf("expected 'test description', got '%s'", i.Description())
	}
	if i.FilterValue() != "test title" {
		t.Errorf("expected 'test title', got '%s'", i.FilterValue())
	}
}

func TestSelectBumpTypeModel(t *testing.T) {
	m := selectBumpTypeModel{
		choice:    "patch",
		quitting:  false,
		cancelled: false,
	}

	// Test Init returns nil
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init to return nil")
	}

	// Test View when not quitting
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view when not quitting")
	}

	// Test View when quitting
	m.quitting = true
	view = m.View()
	if view != "" {
		t.Errorf("expected empty view when quitting, got '%s'", view)
	}
}

func TestConfirmModel(t *testing.T) {
	t.Run("default true", func(t *testing.T) {
		m := confirmModel{
			prompt:       "Test prompt?",
			defaultValue: true,
		}

		view := m.View()
		if view == "" {
			t.Error("expected non-empty view")
		}
		// Should contain [Y/n] for default true
		if len(view) == 0 {
			t.Error("view should not be empty")
		}
	})

	t.Run("default false", func(t *testing.T) {
		m := confirmModel{
			prompt:       "Test prompt?",
			defaultValue: false,
		}

		view := m.View()
		if view == "" {
			t.Error("expected non-empty view")
		}
	})
}

func TestInputMessageModel(t *testing.T) {
	m := inputMessageModel{
		message:   "test message",
		quitting:  false,
		cancelled: false,
	}

	// Test Init returns a command (textarea.Blink)
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init to return a command")
	}

	// Test View when quitting
	m.quitting = true
	view := m.View()
	if view != "" {
		t.Errorf("expected empty view when quitting, got '%s'", view)
	}
}

func TestSelectPreReleaseTypeModel(t *testing.T) {
	options := []PreReleaseTypeOption{
		{Type: semver.PreReleaseAlpha, NewVersion: "v1.0.1-alpha-1"},
		{Type: semver.PreReleaseBeta, NewVersion: "v1.0.1-beta-1"},
	}

	m := selectPreReleaseTypeModel{
		options:   options,
		choice:    semver.PreReleaseAlpha,
		quitting:  false,
		cancelled: false,
	}

	// Test Init returns nil
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init to return nil")
	}

	// Test View when not quitting
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view when not quitting")
	}

	// Test View when quitting
	m.quitting = true
	view = m.View()
	if view != "" {
		t.Errorf("expected empty view when quitting, got '%s'", view)
	}
}

func TestSelectPreReleaseActionModel(t *testing.T) {
	options := []PreReleaseActionOption{
		{Action: PreReleaseActionBump, NewVersion: "v1.0.1-alpha-2"},
		{Action: PreReleaseActionStable, NewVersion: "v1.0.1"},
	}

	m := selectPreReleaseActionModel{
		options:   options,
		choice:    PreReleaseActionBump,
		quitting:  false,
		cancelled: false,
	}

	// Test Init returns nil
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init to return nil")
	}

	// Test View when not quitting
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view when not quitting")
	}

	// Test View when quitting
	m.quitting = true
	view = m.View()
	if view != "" {
		t.Errorf("expected empty view when quitting, got '%s'", view)
	}
}
