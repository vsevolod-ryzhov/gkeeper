package models

import (
	"fmt"
	"gkeeper/internal/tui/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// DashboardModel is the Bubble Tea model for the authenticated user's dashboard.
type DashboardModel struct {
	choices  []string
	cursor   int
	Selected string
	Logout   bool
	Email    string
}

// NewDashboardModel creates a new DashboardModel with default choices.
func NewDashboardModel() DashboardModel {
	return DashboardModel{
		choices:  []string{"list", "new", "logout"},
		cursor:   0,
		Selected: "",
		Logout:   false,
		Email:    "",
	}
}

// Init returns the initial command for the dashboard model.
func (m DashboardModel) Init() tea.Cmd {
	return nil
}

// Update handles keyboard input for dashboard navigation and selection.
func (m DashboardModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
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

// View renders the dashboard screen.
func (m DashboardModel) View() string {
	var s strings.Builder
	s.WriteString(styles.RenderTitle(fmt.Sprintf("Welcome back %s!", m.Email)))
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
