package models

import (
	"context"
	"fmt"
	pb "gkeeper/api/proto"
	"gkeeper/internal/grpcclient"
	"gkeeper/internal/tui/styles"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FileDownloadedMsg is the message sent after a binary secret download attempt.
type FileDownloadedMsg struct {
	Path  string
	Error error
}

// DownloadModel is the Bubble Tea model for downloading binary secrets to disk.
type DownloadModel struct {
	secret     *pb.Secret
	pathInput  textinput.Model
	Back       bool
	ErrorMsg   string
	SuccessMsg string
	client     *grpcclient.Client
}

// NewDownloadModel creates a new DownloadModel for the given binary secret.
func NewDownloadModel(secret *pb.Secret, client *grpcclient.Client) DownloadModel {
	pathInput := textinput.New()
	pathInput.Placeholder = "Enter path to save file"
	pathInput.CharLimit = 500
	pathInput.Width = 50
	pathInput.Focus()

	if fileName := secret.GetFilePath(); fileName != "" {
		pathInput.SetValue(fileName)
	}

	return DownloadModel{
		secret:    secret,
		pathInput: pathInput,
		client:    client,
	}
}

// Init focuses the path input field.
func (m DownloadModel) Init() tea.Cmd {
	return m.pathInput.Focus()
}

// Update handles keyboard input for the download form.
func (m DownloadModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case FileDownloadedMsg:
		if msg.Error != nil {
			m.ErrorMsg = msg.Error.Error()
			m.SuccessMsg = ""
		} else {
			m.SuccessMsg = fmt.Sprintf("File saved to: %s", msg.Path)
			m.ErrorMsg = ""
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.Back = true
			return m, nil
		case "enter":
			savePath := strings.TrimSpace(m.pathInput.Value())
			if savePath == "" {
				m.ErrorMsg = "Save path is required"
				return m, nil
			}
			return m, m.downloadFile(savePath)
		}
	}

	var cmd tea.Cmd
	m.pathInput, cmd = m.pathInput.Update(message)
	return m, cmd
}

func (m *DownloadModel) downloadFile(savePath string) tea.Cmd {
	return func() tea.Msg {
		secret, err := m.client.GetSecret(context.Background(), m.secret.GetId())
		if err != nil {
			return FileDownloadedMsg{Error: fmt.Errorf("failed to fetch secret: %w", err)}
		}

		if err := m.client.DecryptBinarySecret(secret.GetEncryptedData(), savePath); err != nil {
			return FileDownloadedMsg{Error: err}
		}
		return FileDownloadedMsg{Path: savePath}
	}
}

// View renders the download form screen.
func (m DownloadModel) View() string {
	var s strings.Builder

	s.WriteString(styles.RenderTitle("Download Binary Secret"))
	s.WriteString("\n\n")

	s.WriteString(styles.SubtitleStyle.Render(fmt.Sprintf("Title: %s", m.secret.GetTitle())))
	s.WriteString("\n")
	if fileName := m.secret.GetFilePath(); fileName != "" {
		s.WriteString(styles.NormalStyle.Render(fmt.Sprintf("Original file: %s", fileName)))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	s.WriteString(styles.NormalStyle.Render("Save to:"))
	s.WriteString("\n")
	s.WriteString(m.pathInput.View())
	s.WriteString("\n\n")

	if m.ErrorMsg != "" {
		s.WriteString(styles.ErrorStyle.Render("✗ " + m.ErrorMsg))
		s.WriteString("\n\n")
	}

	if m.SuccessMsg != "" {
		s.WriteString(styles.SubtitleStyle.Render("✓ " + m.SuccessMsg))
		s.WriteString("\n\n")
	}

	s.WriteString(styles.DividerStyle.Render(strings.Repeat("─", 50)))
	s.WriteString("\n")
	s.WriteString(styles.FooterStyle.Render("Enter: download • Esc: back • Ctrl+C: quit"))

	return s.String()
}
