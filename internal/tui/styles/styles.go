package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	Highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7B56F9"}
	Text      = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
	Warning   = lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF4444"}
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			PaddingLeft(2).
			PaddingRight(2).
			PaddingTop(1).
			PaddingBottom(1).
			MarginTop(1).
			MarginBottom(1).
			Width(50).
			Align(lipgloss.Center)

	MenuItemStyle = lipgloss.NewStyle().
			PaddingLeft(4).
			PaddingRight(4).
			PaddingTop(1).
			PaddingBottom(1)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				PaddingLeft(4).
				PaddingRight(4).
				PaddingTop(1).
				PaddingBottom(1).
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("230")).
				BorderLeft(true)

	FooterStyle = lipgloss.NewStyle().
			Foreground(Subtle).
			Italic(true).
			MarginTop(1)

	DividerStyle = lipgloss.NewStyle().
			Foreground(Subtle).
			Width(50).
			Align(lipgloss.Left)

	NormalStyle = lipgloss.NewStyle().
			Foreground(Text)

	ActiveInputStyle = lipgloss.NewStyle().
				BorderForeground(Highlight).
				BorderStyle(lipgloss.RoundedBorder()).
				Padding(0, 1)

	InactiveInputStyle = lipgloss.NewStyle().
				BorderForeground(Subtle).
				BorderStyle(lipgloss.RoundedBorder()).
				Padding(0, 1)

	ButtonStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("63")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 3).
			MarginTop(1).
			MarginBottom(1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63"))

	ActiveButtonStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("205")).
				Foreground(lipgloss.Color("230")).
				Padding(0, 3).
				MarginTop(1).
				MarginBottom(1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("205")).
				Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true).
			MarginTop(1)

	FocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
)

func RenderTitle(text string) string {
	return TitleStyle.Render(text)
}

func RenderMenuItem(text string, isSelected bool) string {
	if isSelected {
		return SelectedMenuItemStyle.Render("→ " + text)
	}
	return MenuItemStyle.Render("  " + text)
}

func RenderInputField(input string, placeholder string, isActive bool) string {
	style := InactiveInputStyle
	if isActive {
		style = ActiveInputStyle
	}

	if input == "" {
		return style.Render(lipgloss.NewStyle().Foreground(Subtle).Render(placeholder))
	}
	return style.Render(input)
}

func RenderButton(text string, isActive bool) string {
	if isActive {
		return ActiveButtonStyle.Render(text)
	}
	return ButtonStyle.Render(text)
}
