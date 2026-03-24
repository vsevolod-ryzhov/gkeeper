package models

import (
	"context"
	"fmt"
	"gkeeper/internal/grpcclient"
	"gkeeper/internal/model"
	"gkeeper/internal/tui/styles"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FieldType int

const (
	FieldTypeText FieldType = iota
	FieldTypePassword
	FieldTypeTextArea
)

type Field struct {
	Label      string
	Type       FieldType
	Input      textinput.Model
	TextArea   textarea.Model
	Required   bool
	Validation func(string) error
}

type SecretFormModel struct {
	SecretType   string
	TitleInput   textinput.Model
	Fields       []Field
	Metadata     map[string]string
	CurrentField int
	AddingMeta   bool
	MetaKey      textinput.Model
	MetaValue    textinput.Model
	ErrorMsg     string
	Success      bool
	Back         bool
	Editing      bool
	SecretID     string
	AuthToken    string
	client       *grpcclient.Client
}

func NewSecretFormModel(secretType string, editing bool, existingData *model.Secret, authToken string, client *grpcclient.Client) SecretFormModel {
	m := SecretFormModel{
		SecretType:   secretType,
		Fields:       []Field{},
		Metadata:     make(map[string]string),
		AddingMeta:   false,
		Editing:      editing,
		AuthToken:    authToken,
		CurrentField: 0,
		client:       client,
	}

	m.TitleInput = textinput.New()
	m.TitleInput.Placeholder = "Enter title"
	m.TitleInput.CharLimit = 200
	m.TitleInput.Width = 50
	m.TitleInput.Focus()

	m.MetaKey = textinput.New()
	m.MetaKey.Placeholder = "Key (e.g., website, note, etc.)"
	m.MetaKey.CharLimit = 100
	m.MetaKey.Width = 40

	m.MetaValue = textinput.New()
	m.MetaValue.Placeholder = "Value"
	m.MetaValue.CharLimit = 500
	m.MetaValue.Width = 40

	switch secretType {
	case model.SecretTypeCredentials:
		m.addField("Username", FieldTypeText, true, nil)
		m.addField("Password", FieldTypePassword, true, nil)
		m.addField("URL", FieldTypeText, false, nil)
		m.addField("Notes", FieldTypeTextArea, false, nil)

	case model.SecretTypeText:
		m.addField("Content", FieldTypeTextArea, true, nil)
		m.addField("Notes", FieldTypeTextArea, false, nil)

	case model.SecretTypeCard:
		m.addField("Card Number", FieldTypeText, true, validateCardNumber)
		m.addField("Card Holder Name", FieldTypeText, true, nil)
		m.addField("Expiry Date (MM/YY)", FieldTypeText, true, validateExpiryDate)
		m.addField("CVV", FieldTypePassword, true, validateCVV)
		m.addField("Notes", FieldTypeTextArea, false, nil)

	case model.SecretTypeBinary:
		m.addField("File Path", FieldTypeText, true, nil)
		m.addField("Description", FieldTypeTextArea, false, nil)
	}

	if len(m.Fields) > 0 {
		m.CurrentField = 0
		m.updateFocus()
	}

	if editing && existingData != nil {
		m.SecretID = existingData.ID.String()
	}

	return m
}

func (m *SecretFormModel) addField(label string, fieldType FieldType, required bool, validation func(string) error) {
	var field Field

	switch fieldType {
	case FieldTypeText, FieldTypePassword:
		input := textinput.New()
		input.Placeholder = fmt.Sprintf("Enter %s", strings.ToLower(label))
		if fieldType == FieldTypePassword {
			input.EchoMode = textinput.EchoPassword
			input.EchoCharacter = '•'
		}
		input.CharLimit = 200
		input.Width = 40

		field = Field{
			Label:      label,
			Type:       fieldType,
			Input:      input,
			Required:   required,
			Validation: validation,
		}
	case FieldTypeTextArea:
		ta := textarea.New()
		ta.Placeholder = fmt.Sprintf("Enter %s...", strings.ToLower(label))
		ta.CharLimit = 1000
		ta.SetWidth(50)
		ta.SetHeight(5)

		field = Field{
			Label:    label,
			Type:     fieldType,
			TextArea: ta,
			Required: required,
		}
	}

	m.Fields = append(m.Fields, field)
}

// SetFieldValues populates form fields from a key-value map.
// Keys match those used in collectData (e.g. "username", "password", "url", etc.).
func (m *SecretFormModel) SetFieldValues(data map[string]string) {
	fieldKeyMap := map[string]string{
		"Username":            "username",
		"Password":            "password",
		"URL":                 "url",
		"Content":             "content",
		"Card Number":         "card_number",
		"Card Holder Name":    "card_holder_name",
		"Expiry Date (MM/YY)": "expiry_date",
		"CVV":                 "cvv",
		"File Path":           "file_path",
		"Notes":               "notes",
		"Description":         "notes",
	}

	for i := range m.Fields {
		dataKey, ok := fieldKeyMap[m.Fields[i].Label]
		if !ok {
			continue
		}
		value, exists := data[dataKey]
		if !exists {
			continue
		}
		if m.Fields[i].Type == FieldTypeText || m.Fields[i].Type == FieldTypePassword {
			m.Fields[i].Input.SetValue(value)
		} else if m.Fields[i].Type == FieldTypeTextArea {
			m.Fields[i].TextArea.SetValue(value)
		}
	}
}

func (m SecretFormModel) Init() tea.Cmd {
	return m.TitleInput.Focus()
}

func (m SecretFormModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if m.AddingMeta {
		return m.updateAddingMeta(message)
	}

	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.Back = true
			return m, nil
		case "up", "shift+tab":
			if m.CurrentField > 0 {
				m.CurrentField--
				m.updateFocus()
			}
			return m, nil
		case "down", "tab":
			totalFields := 1 + len(m.Fields) + 1
			if m.CurrentField < totalFields-1 {
				m.CurrentField++
				m.updateFocus()
			}
			return m, nil
		case "enter":
			fieldIndex := m.CurrentField - 1
			if fieldIndex >= 0 && fieldIndex < len(m.Fields) && m.Fields[fieldIndex].Type == FieldTypeTextArea {
				// Skip for Notes textarea multi lining
				break
			}

			if m.CurrentField == 1+len(m.Fields) {
				if m.validateForm() {
					return m, m.saveSecret()
				}
			}
			return m, nil
		case "ctrl+p":
			fieldIndex := m.CurrentField - 1
			if fieldIndex >= 0 && fieldIndex < len(m.Fields) && m.Fields[fieldIndex].Type == FieldTypePassword {
				if m.Fields[fieldIndex].Input.EchoMode == textinput.EchoPassword {
					m.Fields[fieldIndex].Input.EchoMode = textinput.EchoNormal
				} else {
					m.Fields[fieldIndex].Input.EchoMode = textinput.EchoPassword
				}
			}
			return m, nil
		case "ctrl+b":
			m.AddingMeta = true
			m.MetaKey.Focus()
			m.MetaValue.Blur()
			return m, nil
		}
	}

	if m.CurrentField == 0 {
		var cmd tea.Cmd
		m.TitleInput, cmd = m.TitleInput.Update(message)
		return m, cmd
	} else if m.CurrentField <= len(m.Fields) {
		fieldIndex := m.CurrentField - 1
		field := &m.Fields[fieldIndex]
		var cmd tea.Cmd

		if field.Type == FieldTypeText || field.Type == FieldTypePassword {
			field.Input, cmd = field.Input.Update(message)
			return m, cmd
		} else if field.Type == FieldTypeTextArea {
			field.TextArea, cmd = field.TextArea.Update(message)
			return m, cmd
		}
	}

	return m, nil
}

func (m SecretFormModel) updateAddingMeta(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			if msg.String() == "tab" {
				if m.MetaKey.Focused() {
					m.MetaKey.Blur()
					m.MetaValue.Focus()
				} else {
					m.MetaValue.Blur()
					m.MetaKey.Focus()
				}
			} else if msg.String() == "shift+tab" {
				if m.MetaKey.Focused() {
					m.MetaKey.Blur()
					m.MetaValue.Focus()
				} else {
					m.MetaValue.Blur()
					m.MetaKey.Focus()
				}
			}
			return m, nil
		case "up", "down":
			if m.MetaKey.Focused() {
				m.MetaKey.Blur()
				m.MetaValue.Focus()
			} else {
				m.MetaValue.Blur()
				m.MetaKey.Focus()
			}
			return m, nil
		case "enter":
			key := strings.TrimSpace(m.MetaKey.Value())
			value := strings.TrimSpace(m.MetaValue.Value())
			if key != "" {
				m.Metadata[key] = value
			}
			m.AddingMeta = false
			m.MetaKey.Reset()
			m.MetaValue.Reset()
			m.updateFocus()
			return m, nil
		case "esc":
			m.AddingMeta = false
			m.MetaKey.Reset()
			m.MetaValue.Reset()
			m.updateFocus()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.MetaKey, cmd = m.MetaKey.Update(message)
	cmds = append(cmds, cmd)
	m.MetaValue, cmd = m.MetaValue.Update(message)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *SecretFormModel) updateFocus() {
	m.TitleInput.Blur()
	for i := range m.Fields {
		if m.Fields[i].Type == FieldTypeText || m.Fields[i].Type == FieldTypePassword {
			m.Fields[i].Input.Blur()
		} else if m.Fields[i].Type == FieldTypeTextArea {
			m.Fields[i].TextArea.Blur()
		}
	}

	if m.CurrentField == 0 {
		m.TitleInput.Focus()
	} else if m.CurrentField <= len(m.Fields) {
		fieldIndex := m.CurrentField - 1
		if m.Fields[fieldIndex].Type == FieldTypeText || m.Fields[fieldIndex].Type == FieldTypePassword {
			m.Fields[fieldIndex].Input.Focus()
		} else if m.Fields[fieldIndex].Type == FieldTypeTextArea {
			m.Fields[fieldIndex].TextArea.Focus()
		}
	}
}

func (m *SecretFormModel) validateForm() bool {
	for _, field := range m.Fields {
		if field.Required {
			var value string
			if field.Type == FieldTypeText || field.Type == FieldTypePassword {
				value = strings.TrimSpace(field.Input.Value())
			} else if field.Type == FieldTypeTextArea {
				value = strings.TrimSpace(field.TextArea.Value())
			}

			if value == "" {
				m.ErrorMsg = fmt.Sprintf("%s is required", field.Label)
				return false
			}

			if field.Validation != nil {
				if err := field.Validation(value); err != nil {
					m.ErrorMsg = err.Error()
					return false
				}
			}
		}
	}
	return true
}

func (m *SecretFormModel) saveSecret() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		secretData := m.collectData()

		var err error
		if m.Editing {
			err = m.client.UpdateSecret(ctx, m.AuthToken, m.SecretID, m.TitleInput.Value(), m.SecretType, secretData)
		} else {
			err = m.client.CreateSecret(ctx, m.AuthToken, m.TitleInput.Value(), m.SecretType, secretData)
		}

		if err != nil {
			return SaveSecretMsg{Error: err}
		}
		return SaveSecretMsg{Success: true}
	}
}

func (m *SecretFormModel) collectData() map[string]interface{} {
	data := make(map[string]interface{})

	for _, field := range m.Fields {
		switch field.Label {
		case "Username":
			data["username"] = strings.TrimSpace(field.Input.Value())
		case "Password":
			data["password"] = field.Input.Value()
		case "URL":
			data["url"] = strings.TrimSpace(field.Input.Value())
		case "Content":
			data["content"] = strings.TrimSpace(field.TextArea.Value())
		case "Card Number":
			data["card_number"] = strings.TrimSpace(field.Input.Value())
		case "Card Holder Name":
			data["card_holder_name"] = strings.TrimSpace(field.Input.Value())
		case "Expiry Date (MM/YY)":
			data["expiry_date"] = strings.TrimSpace(field.Input.Value())
		case "CVV":
			data["cvv"] = field.Input.Value()
		case "File Path":
			data["file_path"] = strings.TrimSpace(field.Input.Value())
		case "Notes", "Description":
			if field.Type == FieldTypeTextArea {
				data["notes"] = strings.TrimSpace(field.TextArea.Value())
			}
		}
	}

	if len(m.Metadata) > 0 {
		data["metadata"] = m.Metadata
	}

	return data
}

type SaveSecretMsg struct {
	Success bool
	Error   error
}

func (m SecretFormModel) View() string {
	if m.AddingMeta {
		return m.viewAddMetadata()
	}

	var s strings.Builder
	title := "Create New Secret"
	if m.Editing {
		title = "Edit Secret"
	}
	s.WriteString(styles.RenderTitle(fmt.Sprintf("%s: %s", title, m.SecretType)))
	s.WriteString("\n\n")

	cursor := " "
	if m.CurrentField == 0 {
		cursor = ">"
	}
	s.WriteString(fmt.Sprintf("%s Title*: ", cursor))
	s.WriteString(m.TitleInput.View())
	s.WriteString("\n\n")

	for i, field := range m.Fields {
		cursor := " "
		if m.CurrentField == i+1 {
			cursor = ">"
		}

		required := ""
		if field.Required {
			required = "*"
		}

		s.WriteString(fmt.Sprintf("%s %s%s: ", cursor, field.Label, required))

		switch field.Type {
		case FieldTypeText, FieldTypePassword:
			s.WriteString(field.Input.View())
		case FieldTypeTextArea:
			s.WriteString("\n")
			s.WriteString(field.TextArea.View())
		}

		s.WriteString("\n\n")
	}

	if len(m.Metadata) > 0 {
		s.WriteString(styles.SubtitleStyle.Render("Additional Metadata:"))
		s.WriteString("\n")
		for key, value := range m.Metadata {
			s.WriteString(fmt.Sprintf("  • %s: %s\n", key, value))
		}
		s.WriteString("\n")
	}

	cursor = " "
	if m.CurrentField == 1+len(m.Fields) {
		cursor = ">"
	}
	s.WriteString(fmt.Sprintf("%s %s\n\n", cursor, styles.RenderButton("SAVE", m.CurrentField == 1+len(m.Fields))))

	s.WriteString(styles.FooterStyle.Render("[Ctrl+B] Add metadata field • [Ctrl+P] Toggle password visibility"))
	s.WriteString("\n\n")

	if m.ErrorMsg != "" {
		s.WriteString(styles.ErrorStyle.Render("✗ " + m.ErrorMsg))
		s.WriteString("\n\n")
	}

	s.WriteString(styles.DividerStyle.Render(strings.Repeat("─", 50)))
	s.WriteString("\n")
	s.WriteString(styles.FooterStyle.Render("↑/↓: navigate • Enter: save • Esc: back • Ctrl+B: add metadata • Ctrl+C: quit"))

	return s.String()
}

func (m SecretFormModel) viewAddMetadata() string {
	var s strings.Builder

	s.WriteString(styles.RenderTitle("Add Metadata Field"))
	s.WriteString("\n\n")

	s.WriteString(styles.NormalStyle.Render("Key:"))
	s.WriteString("\n")
	s.WriteString(styles.RenderInputField(m.MetaKey.Value(), m.MetaKey.Placeholder, m.MetaKey.Focused()))
	s.WriteString("\n\n")

	s.WriteString(styles.NormalStyle.Render("Value:"))
	s.WriteString("\n")
	s.WriteString(styles.RenderInputField(m.MetaValue.Value(), m.MetaValue.Placeholder, m.MetaValue.Focused()))
	s.WriteString("\n\n")

	s.WriteString(styles.FooterStyle.Render("Enter to save, Esc to cancel"))

	return s.String()
}

func validateCardNumber(s string) error {
	cleaned := strings.ReplaceAll(s, " ", "")
	if len(cleaned) < 13 || len(cleaned) > 19 {
		return fmt.Errorf("card number must be 13-19 digits")
	}
	return nil
}

func validateExpiryDate(s string) error {
	if len(s) != 5 || s[2] != '/' {
		return fmt.Errorf("expiry date must be in MM/YY format")
	}
	return nil
}

func validateCVV(s string) error {
	if len(s) != 3 && len(s) != 4 {
		return fmt.Errorf("CVV must be 3 or 4 digits")
	}
	return nil
}
