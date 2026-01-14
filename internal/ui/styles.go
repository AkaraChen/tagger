package ui

import "github.com/charmbracelet/lipgloss"

var (
	// TitleStyle 标题样式
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	// SelectedStyle 选中项样式
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	// BorderStyle 边框样式
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)

	// SuccessStyle 成功消息样式
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	// ErrorStyle 错误消息样式
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	// InfoStyle 信息样式
	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	// HelpStyle 帮助文本样式
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	// PromptStyle 提示符样式
	PromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)
)
