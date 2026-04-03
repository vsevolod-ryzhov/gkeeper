package models

import (
	"strings"

	"gkeeper/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// MenuModel is the Bubble Tea model for the main menu screen.
type MenuModel struct {
	choices  []string
	cursor   int
	Selected string
}

// NewMenuModel creates a new MenuModel with default menu choices.
func NewMenuModel() MenuModel {
	return MenuModel{
		choices:  []string{"login", "register", "exit"},
		cursor:   0,
		Selected: "",
	}
}

// Init returns the initial command for the menu model.
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles keyboard input for menu navigation and selection.
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.Selected = m.choices[m.cursor]
			return m, nil
		}
	}
	return m, nil
}

// View renders the menu screen.
func (m MenuModel) View() string {
	var s strings.Builder
	s.WriteString(styles.RenderTitle("Welcome to GKeeper!"))
	s.WriteString("\n\n")

	for i, choice := range m.choices {
		s.WriteString(styles.RenderMenuItem(choice, i == m.cursor))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(styles.DividerStyle.Render(strings.Repeat("─", 30)))
	s.WriteString("\n")
	s.WriteString(styles.FooterStyle.Render("↑ up • ↓ down • Enter select • Ctrl+C quit"))

	return s.String()
}
