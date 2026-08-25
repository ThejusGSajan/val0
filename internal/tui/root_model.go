package tui

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/auth"
	"github.com/val-tracker/val-tracker/internal/cache"
	"github.com/val-tracker/val-tracker/internal/models"
	"github.com/val-tracker/val-tracker/internal/sprite"
)

type RootState int

const (
	StateAuthenticating RootState = iota
	StateAuthError
	StateMain
)

type AuthSuccessMsg struct {
	Session *models.Session
}

type AuthFailedMsg struct {
	Err error
}

type RetryAuthMsg struct{}

type RootModel struct {
	state      RootState
	errorModel ErrorModel
	mainModel  MainModel
	width      int
	height     int
}

func NewRootModel() RootModel {
	return RootModel{
		state: StateAuthenticating,
	}
}

func (m RootModel) Init() tea.Cmd {
	return authenticateCmd
}

func authenticateCmd() tea.Msg {
	session, err := auth.BuildSession()
	if err != nil {
		return AuthFailedMsg{Err: err}
	}

	ver, err := cache.FetchRemoteVersion()
	if err != nil {
		session.ClientVersion = "release-13.02-shipping-17-5277781"
	} else {
		session.ClientVersion = ver.RiotClientVersion
	}

	return AuthSuccessMsg{Session: session}
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.errorModel.SetSize(msg.Width, msg.Height)
		if m.state == StateMain {
			var cmd tea.Cmd
			var model tea.Model
			model, cmd = m.mainModel.Update(msg)
			if mm, ok := model.(MainModel); ok {
				m.mainModel = mm
			}
			return m, cmd
		}
		return m, nil

	case AuthSuccessMsg:
		m.state = StateMain
		m.mainModel = NewMainModel(msg.Session)
		if m.width > 0 && m.height > 0 {
			var model tea.Model
			model, _ = m.mainModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			if mm, ok := model.(MainModel); ok {
				m.mainModel = mm
			}
		}
		return m, m.mainModel.Init()

	case AuthFailedMsg:
		m.state = StateAuthError
		errMsg := fmt.Sprintf("Authentication failed:\n%s", msg.Err.Error())
		if errors.Is(msg.Err, auth.ErrLockfileNotFound) {
			errMsg = "Please start the Riot Client first.\n(Lockfile was not found)"
		}
		m.errorModel = NewErrorModel(errMsg)
		m.errorModel.SetSize(m.width, m.height)
		return m, nil

	case RetryAuthMsg:
		m.state = StateAuthenticating
		return m, authenticateCmd

	case tea.KeyMsg:
		if m.state == StateAuthError {
			switch msg.String() {
			case "r", "R":
				m.state = StateAuthenticating
				return m, authenticateCmd
			case "q", "ctrl+c", "esc":
				return m, tea.Quit
			}
		} else if m.state == StateAuthenticating {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
	}

	if m.state == StateMain {
		var cmd tea.Cmd
		var model tea.Model
		model, cmd = m.mainModel.Update(msg)
		if mm, ok := model.(MainModel); ok {
			m.mainModel = mm
		}
		return m, cmd
	}

	return m, nil
}

func (m RootModel) View() string {
	var finalScreen string
	switch m.state {
	case StateAuthenticating:
		content := lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Render("⟳  Connecting to Riot Client...")
		if m.width > 0 && m.height > 0 {
			finalScreen = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
		} else {
			finalScreen = "\n\n  " + content + "\n"
		}

	case StateAuthError:
		finalScreen = m.errorModel.View()

	case StateMain:
		finalScreen = m.mainModel.View()
	}

	return sprite.InjectPayloads(finalScreen)
}
