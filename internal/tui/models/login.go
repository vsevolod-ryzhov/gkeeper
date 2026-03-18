package models

import (
	"context"
	"gkeeper/internal/grpcclient"
	"gkeeper/internal/tui/styles"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
)

type LoginModel struct {
	emailInput textinput.Model
	passInput  textinput.Model
	focusIndex int
	Success    bool
	Back       bool
	Email      string
	ErrorMsg   string
}

func NewLoginModel() LoginModel {
	email := textinput.New()
	email.Placeholder = "Email"
	email.Focus()
	email.CharLimit = 100
	email.Width = 30

	pass := textinput.New()
	pass.Placeholder = "Password"
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = 'x'
	pass.CharLimit = 100
	pass.Width = 30

	return LoginModel{
		emailInput: email,
		passInput:  pass,
		focusIndex: 0,
		Success:    false,
		Back:       false,
	}
}

func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			m.Back = true
			return m, nil

		case "tab", "shift+tab", "enter", "up", "down":
			m.ErrorMsg = ""
			if msg.String() == "enter" && m.focusIndex == 2 {
				if m.validateForm() {
					ctx := context.Background()
					client := grpcclient.NewClient(&zap.Logger{})
					defer client.Close()
					err := client.Login(ctx, m.emailInput.Value(), m.passInput.Value())
					if err != nil {
						m.ErrorMsg = "Invalid email or password"
					} else {
						m.Success = true
						m.Email = m.emailInput.Value()
					}
				} else {
					m.ErrorMsg = "Invalid email or password"
				}
				return m, nil
			}

			// cyclic switching between inputs
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex < 0 {
				m.focusIndex = 2
			} else if m.focusIndex > 2 {
				m.focusIndex = 0
			}

			cmds = append(cmds, m.updateFocus())
		}
	}

	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.emailInput, cmd = m.emailInput.Update(message)
		cmds = append(cmds, cmd)
	} else if m.focusIndex == 1 {
		m.passInput, cmd = m.passInput.Update(message)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *LoginModel) updateFocus() tea.Cmd {
	if m.focusIndex == 0 {
		m.emailInput.Focus()
		m.passInput.Blur()
		return textinput.Blink
	} else if m.focusIndex == 1 {
		m.emailInput.Blur()
		m.passInput.Focus()
		return textinput.Blink
	} else {
		m.emailInput.Blur()
		m.passInput.Blur()
		return nil
	}
}

func (m LoginModel) validateForm() bool {
	email := m.emailInput.Value()
	pass := m.passInput.Value()
	return strings.Contains(email, "@") && len(pass) >= 3
}

func (m LoginModel) View() string {
	var b strings.Builder

	b.WriteString(styles.RenderTitle("Login"))
	b.WriteString("\n")

	b.WriteString(styles.NormalStyle.Render("Email:"))
	b.WriteString("\n")
	b.WriteString(styles.RenderInputField(m.emailInput.Value(), m.emailInput.Placeholder, m.focusIndex == 0))
	b.WriteString("\n\n")

	b.WriteString(styles.NormalStyle.Render("Password:"))
	b.WriteString("\n")

	b.WriteString(styles.RenderInputField(m.passInput.Value(), m.passInput.Placeholder, m.focusIndex == 1))
	b.WriteString("\n\n")

	b.WriteString(styles.RenderButton("LOGIN", m.focusIndex == 2))
	b.WriteString("\n\n")

	if m.ErrorMsg != "" {
		b.WriteString(styles.ErrorStyle.Render("✗ " + m.ErrorMsg))
		b.WriteString("\n")
	}

	b.WriteString(styles.FooterStyle.Render("\nPress Esc to go back, Ctrl+C to quit.\n"))
	return b.String()
}
