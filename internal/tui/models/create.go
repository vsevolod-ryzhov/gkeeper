package models

import (
	"gkeeper/internal/grpcclient"
	"gkeeper/internal/model"
	"gkeeper/internal/tui/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CreateModel is the Bubble Tea model for the secret type selection and creation form.
type CreateModel struct {
	choices      []string
	cursor       int
	Selected     string
	ShowForm     bool
	FormModel    SecretFormModel
	AuthToken    string
	EditComplete bool
	Back         bool
	client       *grpcclient.Client
}

// NewCreateModel creates a new CreateModel with available secret type choices.
func NewCreateModel(client *grpcclient.Client) CreateModel {
	return CreateModel{
		choices:   []string{model.SecretTypeCredentials, model.SecretTypeText, model.SecretTypeCard, model.SecretTypeBinary},
		cursor:    0,
		Selected:  "",
		ShowForm:  false,
		AuthToken: "",
		Back:      false,
		client:    client,
	}
}

// Init returns the initial command for the create model.
func (m CreateModel) Init() tea.Cmd {
	return nil
}

// Update handles keyboard input for secret type selection and delegates to the form model.
func (m CreateModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if m.ShowForm {
		updatedForm, cmd := m.FormModel.Update(message)
		m.FormModel = updatedForm.(SecretFormModel)

		if saveMsg, ok := message.(SaveSecretMsg); ok {
			if saveMsg.Success {
				m.ShowForm = false
				if m.FormModel.Editing {
					m.EditComplete = true
				}
				m.Selected = ""
				return m, nil
			}
			if saveMsg.Error != nil {
				m.FormModel.ErrorMsg = saveMsg.Error.Error()
			}
		}

		if m.FormModel.Success {
			m.ShowForm = false
			m.Selected = ""
			return m, nil
		}

		if m.FormModel.Back {
			m.ShowForm = false
			return m, nil
		}

		return m, cmd
	}

	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.Back = true
			return m, nil
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
			m.FormModel = NewSecretFormModel(m.Selected, false, nil, m.AuthToken, m.client)
			m.ShowForm = true
			return m, m.FormModel.Init()
		}
	}
	return m, nil
}

// View renders the secret type selection or the active form.
func (m CreateModel) View() string {
	if m.ShowForm {
		return m.FormModel.View()
	}

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
	s.WriteString(styles.FooterStyle.Render("↑ up • ↓ down • Enter select • Esc: back • Ctrl+C: quit"))

	return s.String()
}
