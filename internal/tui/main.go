package tui

import (
	"gkeeper/internal/tui/models"

	tea "github.com/charmbracelet/bubbletea"
)

type SessionState int

const (
	MenuView SessionState = iota
	LoginView
	RegisterView
	DashboardView
)

type MainModel struct {
	state     SessionState
	menu      models.MenuModel
	login     models.LoginModel
	register  models.RegisterModel
	dashboard models.DashboardModel

	authToken string
	userEmail string
}

func NewMainModel() MainModel {
	return MainModel{
		state:     MenuView,
		menu:      models.NewMenuModel(),
		login:     models.NewLoginModel(),
		register:  models.NewRegisterModel(),
		dashboard: models.NewDashboardModel(),
	}
}

func (m MainModel) Init() tea.Cmd {
	switch m.state {
	case MenuView:
		return m.menu.Init()
	case LoginView:
		return m.login.Init()
	case RegisterView:
		return m.register.Init()
	case DashboardView:
		return m.dashboard.Init()
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
			case "register":
				m.state = RegisterView
				m.menu.Selected = ""
				cmds = append(cmds, m.register.Init())
			case "dashboard":
				m.state = DashboardView
				m.menu.Selected = ""
				cmds = append(cmds, m.dashboard.Init())
			case "exit":
				return m, tea.Quit
			}
		}

	case LoginView:
		updatedLogin, loginCmd := m.login.Update(msg)
		m.login = updatedLogin.(models.LoginModel)
		cmds = append(cmds, loginCmd)

		if m.login.Success {
			m.authToken = m.login.Token
			m.userEmail = m.login.Email
			m.state = DashboardView
			m.login.Success = false
			cmds = append(cmds, m.dashboard.Init())
		}

		if m.login.Back {
			m.state = MenuView
			m.login.Back = false
		}

	case RegisterView:
		updateRegister, registerCmd := m.register.Update(msg)
		m.register = updateRegister.(models.RegisterModel)
		cmds = append(cmds, registerCmd)

		if m.register.Success {
			m.state = MenuView
			m.register.Success = false
		}

		if m.register.Back {
			m.state = MenuView
			m.register.Back = false
		}

	case DashboardView:
		updatedDashboard, dashboardCmd := m.dashboard.Update(msg)
		m.dashboard = updatedDashboard.(models.DashboardModel)
		m.dashboard.Email = m.userEmail
		cmds = append(cmds, dashboardCmd)

		if m.dashboard.Selected != "" {
			switch m.dashboard.Selected {
			case "logout":
				m.state = MenuView
				m.dashboard.Selected = ""
				cmds = append(cmds, m.menu.Init())
			}
		}

		if m.dashboard.Logout {
			m.state = MenuView
			m.authToken = ""
			m.userEmail = ""
			m.dashboard.Logout = false
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
	case RegisterView:
		return m.register.View()
	case DashboardView:
		return m.dashboard.View()
	default:
		return "Unknown view"
	}
}
