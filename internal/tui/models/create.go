package models

import (
	"gkeeper/internal/model"
	"gkeeper/internal/tui/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type CreateModel struct {
	choices  []string
	cursor   int
	Selected string
}

func NewCreateModel() CreateModel {
	return CreateModel{
		choices:  []string{model.SecretTypeCredentials, model.SecretTypeText, model.SecretTypeCard, model.SecretTypeBinary, "back"},
		cursor:   0,
		Selected: "",
	}
}

func (m CreateModel) Init() tea.Cmd {
	return nil
}

func (m CreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m CreateModel) View() string {
	var s strings.Builder
	s.WriteString(styles.RenderTitle("Create new secret record"))
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
