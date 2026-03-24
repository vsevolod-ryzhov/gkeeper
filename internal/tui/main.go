package tui

import (
	"encoding/json"

	pb "gkeeper/api/proto"
	"gkeeper/internal/grpcclient"
	"gkeeper/internal/tui/models"

	tea "github.com/charmbracelet/bubbletea"
)

type SessionState int

const (
	MenuView SessionState = iota
	LoginView
	RegisterView
	DashboardView
	CreateView
	ListView
)

type MainModel struct {
	state     SessionState
	menu      models.MenuModel
	login     models.LoginModel
	register  models.RegisterModel
	dashboard models.DashboardModel
	create    models.CreateModel
	list      models.ListModel
	client    *grpcclient.Client

	authToken string
	userEmail string
}

func NewMainModel(client *grpcclient.Client) MainModel {
	return MainModel{
		state:     MenuView,
		menu:      models.NewMenuModel(),
		login:     models.NewLoginModel(client),
		register:  models.NewRegisterModel(client),
		dashboard: models.NewDashboardModel(),
		create:    models.NewCreateModel(client),
		list:      models.NewListModel(client),
		client:    client,
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
	case CreateView:
		return m.create.Init()
	case ListView:
		return m.list.Init()
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
			case "new":
				m.state = CreateView
				m.dashboard.Selected = ""
				cmds = append(cmds, m.create.Init())
			case "list":
				m.state = ListView
				m.dashboard.Selected = ""
				m.list.AuthToken = m.authToken
				cmds = append(cmds, m.list.Init())
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

	case CreateView:
		m.create.AuthToken = m.authToken
		updatedCreate, createCmd := m.create.Update(msg)
		m.create = updatedCreate.(models.CreateModel)
		cmds = append(cmds, createCmd)

		if m.create.EditComplete {
			m.create.EditComplete = false
			m.state = ListView
			m.list.AuthToken = m.authToken
			cmds = append(cmds, m.list.Init())
		}

		if m.create.Back {
			m.state = DashboardView
			m.create.Back = false
			cmds = append(cmds, m.dashboard.Init())
		}

	case ListView:
		updatedList, listCmd := m.list.Update(msg)
		m.list = updatedList.(models.ListModel)
		cmds = append(cmds, listCmd)

		if m.list.Back {
			m.state = DashboardView
			m.list.Back = false
			cmds = append(cmds, m.dashboard.Init())
		}

		if m.list.Selected != nil {
			selected := m.list.Selected
			m.list.Selected = nil
			m.create.AuthToken = m.authToken
			m.create.ShowForm = true
			m.create.FormModel = NewEditFormFromProto(selected, m.authToken, m.client)
			m.state = CreateView
			cmds = append(cmds, m.create.FormModel.Init())
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
	case CreateView:
		return m.create.View()
	case ListView:
		return m.list.View()
	default:
		return "Unknown view"
	}
}

func NewEditFormFromProto(secret *pb.Secret, authToken string, client *grpcclient.Client) models.SecretFormModel {
	form := models.NewSecretFormModel(secret.GetType(), true, nil, authToken, client)
	form.SecretID = secret.GetId()
	form.TitleInput.SetValue(secret.GetTitle())

	crypto := client.GetCrypto()
	if crypto != nil && len(secret.GetEncryptedData()) > 0 {
		decrypted, err := crypto.Decrypt(string(secret.GetEncryptedData()))
		if err == nil {
			var data map[string]string
			if json.Unmarshal(decrypted, &data) == nil {
				form.SetFieldValues(data)
			}
		}
	}

	if secret.GetMetadata() != "" {
		var meta map[string]string
		if json.Unmarshal([]byte(secret.GetMetadata()), &meta) == nil {
			for k, v := range meta {
				form.Metadata[k] = v
			}
		}
	}

	return form
}
