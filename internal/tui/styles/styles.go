package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
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
