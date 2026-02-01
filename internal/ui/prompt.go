package ui

import (
	"fmt"
	"strings"

	"github.com/AkaraChen/tagger/internal/semver"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BumpTypeOption 版本更新选项
type BumpTypeOption struct {
	Type       string
	NewVersion string
	Desc       string
}

// SelectBumpType 选择版本更新类型
func SelectBumpType(currentVersion string, options []BumpTypeOption) (string, error) {
	items := make([]list.Item, len(options))
	for i, opt := range options {
		items[i] = item{
			title: opt.Type,
			desc:  fmt.Sprintf("%s → %s (%s)", currentVersion, opt.NewVersion, opt.Desc),
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("Current Version: %s", currentVersion)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Styles.Title = TitleStyle

	m := selectBumpTypeModel{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := finalModel.(selectBumpTypeModel); ok {
		if m.cancelled {
			return "", fmt.Errorf("cancelled")
		}
		return m.choice, nil
	}

	return "", fmt.Errorf("unexpected error")
}

// ConfirmAddMessage 询问是否添加 tag message
func ConfirmAddMessage() (bool, error) {
	m := confirmModel{
		prompt:       "Add a tag message?",
		defaultValue: false,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	if m, ok := finalModel.(confirmModel); ok {
		if m.cancelled {
			return false, fmt.Errorf("cancelled")
		}
		return m.confirmed, nil
	}

	return false, fmt.Errorf("unexpected error")
}

// InputTagMessage 输入 tag message
func InputTagMessage(defaultText string) (string, error) {
	ta := textarea.New()
	ta.Placeholder = "Enter tag message..."
	ta.Focus()
	ta.SetWidth(60)
	ta.SetHeight(5)

	if defaultText != "" {
		ta.SetValue(defaultText)
	}

	m := inputMessageModel{textarea: ta}
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := finalModel.(inputMessageModel); ok {
		if m.cancelled {
			return "", fmt.Errorf("cancelled")
		}
		return m.message, nil
	}

	return "", fmt.Errorf("unexpected error")
}

// ConfirmCreateTag 确认创建 tag
func ConfirmCreateTag(oldVersion, newVersion, message string) (bool, error) {
	prompt := fmt.Sprintf("Create tag %s → %s?", oldVersion, newVersion)
	if message != "" {
		msgPreview := message
		if len(msgPreview) > 50 {
			msgPreview = msgPreview[:50] + "..."
		}
		prompt = fmt.Sprintf("Create tag %s → %s\nMessage: %s", oldVersion, newVersion, msgPreview)
	}

	m := confirmModel{
		prompt:       prompt,
		defaultValue: true,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	if m, ok := finalModel.(confirmModel); ok {
		if m.cancelled {
			return false, fmt.Errorf("cancelled")
		}
		return m.confirmed, nil
	}

	return false, fmt.Errorf("unexpected error")
}

// ConfirmPush 确认推送 tag
func ConfirmPush(version string) (bool, error) {
	m := confirmModel{
		prompt:       fmt.Sprintf("Push tag %s to remote?", version),
		defaultValue: true,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	if m, ok := finalModel.(confirmModel); ok {
		if m.cancelled {
			return false, fmt.Errorf("cancelled")
		}
		return m.confirmed, nil
	}

	return false, fmt.Errorf("unexpected error")
}

// ConfirmOpenRepo 确认打开 GitHub 仓库
func ConfirmOpenRepo() (bool, error) {
	m := confirmModel{
		prompt:       "Open GitHub repository in browser?",
		defaultValue: false,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	if m, ok := finalModel.(confirmModel); ok {
		if m.cancelled {
			return false, fmt.Errorf("cancelled")
		}
		return m.confirmed, nil
	}

	return false, fmt.Errorf("unexpected error")
}

// ConfirmPreRelease asks if the user wants to create a pre-release
func ConfirmPreRelease() (bool, error) {
	m := confirmModel{
		prompt:       "Is this a pre-release?",
		defaultValue: false,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	if m, ok := finalModel.(confirmModel); ok {
		if m.cancelled {
			return false, fmt.Errorf("cancelled")
		}
		return m.confirmed, nil
	}

	return false, fmt.Errorf("unexpected error")
}

// PreReleaseTypeOption represents an option for pre-release type selection
type PreReleaseTypeOption struct {
	Type       semver.PreReleaseType
	NewVersion string
	LatestInfo string
}

// SelectPreReleaseType allows user to select pre-release type (alpha/beta/rc)
func SelectPreReleaseType(currentVersion string, options []PreReleaseTypeOption) (semver.PreReleaseType, error) {
	items := make([]list.Item, len(options))
	for i, opt := range options {
		desc := fmt.Sprintf("%s → %s", currentVersion, opt.NewVersion)
		if opt.LatestInfo != "" {
			desc += fmt.Sprintf(" (latest: %s)", opt.LatestInfo)
		} else {
			desc += " (none exists)"
		}
		items[i] = item{
			title: string(opt.Type),
			desc:  desc,
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select pre-release type"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Styles.Title = TitleStyle

	m := selectPreReleaseTypeModel{list: l, options: options}
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := finalModel.(selectPreReleaseTypeModel); ok {
		if m.cancelled {
			return "", fmt.Errorf("cancelled")
		}
		return m.choice, nil
	}

	return "", fmt.Errorf("unexpected error")
}

// PreReleaseAction represents an action for existing pre-release versions
type PreReleaseAction string

const (
	PreReleaseActionBump    PreReleaseAction = "bump"
	PreReleaseActionAdvance PreReleaseAction = "advance"
	PreReleaseActionStable  PreReleaseAction = "stable"
)

// PreReleaseActionOption represents an option for pre-release action selection
type PreReleaseActionOption struct {
	Action     PreReleaseAction
	NewVersion string
	Desc       string
}

// SelectPreReleaseAction allows user to select action for existing pre-release
func SelectPreReleaseAction(currentVersion string, options []PreReleaseActionOption) (PreReleaseAction, error) {
	items := make([]list.Item, len(options))
	for i, opt := range options {
		desc := fmt.Sprintf("%s → %s", currentVersion, opt.NewVersion)
		if opt.Desc != "" {
			desc += fmt.Sprintf(" (%s)", opt.Desc)
		}
		items[i] = item{
			title: string(opt.Action),
			desc:  desc,
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("Current pre-release: %s", currentVersion)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Styles.Title = TitleStyle

	m := selectPreReleaseActionModel{list: l, options: options}
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := finalModel.(selectPreReleaseActionModel); ok {
		if m.cancelled {
			return "", fmt.Errorf("cancelled")
		}
		return m.choice, nil
	}

	return "", fmt.Errorf("unexpected error")
}

// --- Models ---

// item 实现 list.Item 接口
type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// selectBumpTypeModel 选择版本更新类型的 Model
type selectBumpTypeModel struct {
	list      list.Model
	choice    string
	quitting  bool
	cancelled bool
}

func (m selectBumpTypeModel) Init() tea.Cmd {
	return nil
}

func (m selectBumpTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.choice = i.title
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectBumpTypeModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

// confirmModel 确认的 Model
type confirmModel struct {
	prompt       string
	defaultValue bool
	confirmed    bool
	cancelled    bool
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "y", "Y":
			m.confirmed = true
			return m, tea.Quit

		case "n", "N":
			m.confirmed = false
			return m, tea.Quit

		case "enter":
			m.confirmed = m.defaultValue
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m confirmModel) View() string {
	defaultIndicator := "[y/N]"
	if m.defaultValue {
		defaultIndicator = "[Y/n]"
	}

	return fmt.Sprintf("\n%s %s ",
		InfoStyle.Render(m.prompt),
		HelpStyle.Render(defaultIndicator),
	)
}

// inputMessageModel Tag Message 输入的 Model
type inputMessageModel struct {
	textarea  textarea.Model
	message   string
	quitting  bool
	cancelled bool
}

func (m inputMessageModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m inputMessageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			m.quitting = true
			return m, tea.Quit

		case tea.KeyCtrlD:
			m.message = strings.TrimSpace(m.textarea.Value())
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m inputMessageModel) View() string {
	if m.quitting {
		return ""
	}

	help := HelpStyle.Render("Ctrl+D to finish • Esc to cancel")
	charCount := HelpStyle.Render(fmt.Sprintf("%d characters", len(m.textarea.Value())))

	content := TitleStyle.Render("Tag Message") + "\n\n" +
		m.textarea.View() + "\n\n" +
		help + " • " + charCount

	return "\n" + BorderStyle.Render(content) + "\n"
}

// selectPreReleaseTypeModel 选择预发布类型的 Model
type selectPreReleaseTypeModel struct {
	list      list.Model
	options   []PreReleaseTypeOption
	choice    semver.PreReleaseType
	quitting  bool
	cancelled bool
}

func (m selectPreReleaseTypeModel) Init() tea.Cmd {
	return nil
}

func (m selectPreReleaseTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			m.quitting = true
			return m, tea.Quit

		case "enter":
			idx := m.list.Index()
			if idx >= 0 && idx < len(m.options) {
				m.choice = m.options[idx].Type
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectPreReleaseTypeModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

// selectPreReleaseActionModel 选择预发布操作的 Model
type selectPreReleaseActionModel struct {
	list      list.Model
	options   []PreReleaseActionOption
	choice    PreReleaseAction
	quitting  bool
	cancelled bool
}

func (m selectPreReleaseActionModel) Init() tea.Cmd {
	return nil
}

func (m selectPreReleaseActionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			m.quitting = true
			return m, tea.Quit

		case "enter":
			idx := m.list.Index()
			if idx >= 0 && idx < len(m.options) {
				m.choice = m.options[idx].Action
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectPreReleaseActionModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}
