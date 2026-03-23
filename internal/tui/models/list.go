package models

import (
	"context"
	"fmt"
	pb "gkeeper/api/proto"
	"gkeeper/internal/grpcclient"
	"gkeeper/internal/tui/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type SecretsLoadedMsg struct {
	Secrets []*pb.Secret
	Error   error
}

type ListModel struct {
	secrets   []*pb.Secret
	cursor    int
	Selected  *pb.Secret
	Back      bool
	ErrorMsg  string
	Loading   bool
	AuthToken string
	client    *grpcclient.Client
}

func NewListModel(client *grpcclient.Client) ListModel {
	return ListModel{
		client:  client,
		Loading: true,
	}
}

func (m ListModel) Init() tea.Cmd {
	return m.loadSecrets()
}

func (m ListModel) loadSecrets() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		secrets, err := m.client.GetSecrets(ctx, m.AuthToken)
		if err != nil {
			return SecretsLoadedMsg{Error: err}
		}
		return SecretsLoadedMsg{Secrets: secrets}
	}
}

func (m ListModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case SecretsLoadedMsg:
		m.Loading = false
		if msg.Error != nil {
			m.ErrorMsg = msg.Error.Error()
			return m, nil
		}
		m.secrets = msg.Secrets
		m.cursor = 0
		return m, nil

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
			if m.cursor < len(m.secrets)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.secrets) > 0 {
				m.Selected = m.secrets[m.cursor]
			}
			return m, nil
		}
	}
	return m, nil
}

func (m ListModel) View() string {
	var s strings.Builder

	s.WriteString(styles.RenderTitle("My Secrets"))
	s.WriteString("\n\n")

	if m.Loading {
		s.WriteString(styles.NormalStyle.Render("Loading..."))
		s.WriteString("\n")
		return s.String()
	}

	if m.ErrorMsg != "" {
		s.WriteString(styles.ErrorStyle.Render("✗ " + m.ErrorMsg))
		s.WriteString("\n\n")
	}

	if len(m.secrets) == 0 {
		s.WriteString(styles.NormalStyle.Render("No secrets found."))
		s.WriteString("\n\n")
		s.WriteString(styles.FooterStyle.Render("Press Esc to go back"))
		return s.String()
	}

	for i, secret := range m.secrets {
		label := fmt.Sprintf("[%s] %s", secret.GetType(), secret.GetTitle())
		s.WriteString(styles.RenderMenuItem(label, i == m.cursor))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(styles.DividerStyle.Render(strings.Repeat("─", 40)))
	s.WriteString("\n")
	s.WriteString(styles.FooterStyle.Render("↑/↓: navigate • Enter: edit • Esc: back • Ctrl+C: quit"))

	return s.String()
}
