package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectBumpType 选择版本更新类型
func SelectBumpType(currentVersion, patchVersion, minorVersion, majorVersion string) (string, error) {
	items := []list.Item{
		item{
			title: "patch",
			desc:  fmt.Sprintf("%s → %s (补丁更新)", currentVersion, patchVersion),
		},
		item{
			title: "minor",
			desc:  fmt.Sprintf("%s → %s (小版本更新)", currentVersion, minorVersion),
		},
		item{
			title: "major",
			desc:  fmt.Sprintf("%s → %s (大版本更新)", currentVersion, majorVersion),
		},
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
