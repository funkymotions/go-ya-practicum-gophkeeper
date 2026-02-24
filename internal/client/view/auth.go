package view

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/utils"
)

type authViewType string

const (
	AuthViewType     authViewType = "auth"
	RegisterViewType authViewType = "register"
)

type authModel struct {
	title     string
	viewType  authViewType
	inputs    []textinput.Model
	PrevModel types.NamedTeaModel
	focused   int
	service   ports.ClientAuthService
	err       error
	state     *types.State
}

type AuthModelArgs struct {
	Service  ports.ClientAuthService
	ViewType authViewType
	State    *types.State
}

func NewAuthModel(args AuthModelArgs) *authModel {
	usernameTextInput := textinput.New()
	passwordTextInput := textinput.New()
	usernameTextInput.Placeholder = "Username"
	usernameTextInput.CharLimit = 32
	usernameTextInput.Width = 20
	usernameTextInput.Focus()

	passwordTextInput.Placeholder = "Password"
	passwordTextInput.EchoMode = textinput.EchoPassword
	passwordTextInput.EchoCharacter = '•'
	passwordTextInput.CharLimit = 32
	passwordTextInput.Width = 20

	var title string
	switch args.ViewType {
	case RegisterViewType:
		title = "Registration"
	case AuthViewType:
		title = "Authorization"
	}

	return &authModel{
		inputs:   []textinput.Model{usernameTextInput, passwordTextInput},
		title:    title,
		viewType: args.ViewType,
		service:  args.Service,
		state:    args.State,
	}
}

func (rm *authModel) Init() tea.Cmd {
	return textinput.Blink
}

func (rm *authModel) GetTitle() string {
	return rm.title
}

func (rm *authModel) SetPrevModel(m types.NamedTeaModel) {
	rm.PrevModel = m
}

func (am *authModel) Reset() {
	for i := range am.inputs {
		am.inputs[i].SetValue("")
	}
	am.focused = 0
	am.err = nil
	am.inputs[am.focused].Focus()
}

func (rm *authModel) IsAuthorizedModel() bool {
	return false
}

func (rm *authModel) submit() error {
	username := rm.inputs[0].Value()
	password := rm.inputs[1].Value()
	if username == "" || password == "" {
		return fmt.Errorf("username and password cannot be empty")
	}

	var token string
	var err error
	if rm.viewType == RegisterViewType {
		token, err = rm.service.Register(username, password)
	}
	if rm.viewType == AuthViewType {
		token, err = rm.service.Authenticate(username, password)
	}
	if err != nil {
		return err
	}

	rm.state.Token = token
	rm.state.IsOnline = true
	claims, err := utils.ParseUnverifiedJWT(token)
	if err != nil {
		return err
	}

	rm.state.UserID = claims.UserID

	return nil
}

func (rm *authModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			rm.Reset()
			return rm.PrevModel, nil
		case tea.KeyEnter:
			if rm.focused >= len(rm.inputs)-1 {
				// TODO: now it's blocking, make it async using goroutines and channels
				err := rm.submit()
				if err != nil {
					rm.err = err

					return rm, nil
				}

				return rm.PrevModel, rm.PrevModel.Init()
			}

			rm.focused++
			rm.inputs[rm.focused].Focus()

			return rm, nil
		case tea.KeyCtrlC:
			return rm, tea.Quit
		}
	}

	rm.inputs[rm.focused], cmd = rm.inputs[rm.focused].Update(msg)

	return rm, cmd
}

func (rm *authModel) View() string {
	var errText string
	if rm.err != nil {
		errText = fmt.Sprintf("Error: %v, press ESC to retry\n", rm.err)
	}
	s := "\n== " + rm.title + " ==\n"
	s += errText
	s += "\n Please enter your credentials \n\n"
	s += rm.inputs[rm.focused].View()
	s += "\n\n(Press 'ENTER' to submit, 'ESC' to go back.)\n"

	return s
}
