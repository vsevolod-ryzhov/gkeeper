package tui

import (
	"gkeeper/internal/tui/models"

	tea "github.com/charmbracelet/bubbletea"
)

type SessionState int

const (
	MenuView SessionState = iota
	LoginView
)

type MainModel struct {
	state SessionState
	menu  models.MenuModel
	login models.LoginModel

	userEmail string
}

func NewMainModel() MainModel {
	return MainModel{
		state: MenuView,
		menu:  models.NewMenuModel(),
		login: models.NewLoginModel(),
	}
}

func (m MainModel) Init() tea.Cmd {
	switch m.state {
	case MenuView:
		return m.menu.Init()
	case LoginView:
		return m.login.Init()
	}

	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch m.state {
	case MenuView:
		updatedMenu, menuCmd := m.menu.Update(msg)
		m.menu = updatedMenu.(models.MenuModel)
		cmds = append(cmds, menuCmd)

		if m.menu.Selected != "" {
			switch m.menu.Selected {
			case "login":
				m.state = LoginView
				m.menu.Selected = ""
				cmds = append(cmds, m.login.Init())
			case "exit":
				return m, tea.Quit
			}
		}

	case LoginView:
		updatedLogin, loginCmd := m.login.Update(msg)
		m.login = updatedLogin.(models.LoginModel)
		cmds = append(cmds, loginCmd)

		if m.login.Success {
			// Display dashboard screen for logged-in user
		}

		if m.login.Back {
			m.state = MenuView
			m.login.Back = false
		}
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	switch m.state {
	case MenuView:
		return m.menu.View()
	case LoginView:
		return m.login.View()
	default:
		return "Unknown view"
	}
}
