package client

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/infrastructure/grpc"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/auth"
	"google.golang.org/protobuf/proto"
)

type authViewType string

const (
	AuthView     authViewType = "auth"
	RegisterView authViewType = "register"
)

type authModel struct {
	title      string
	viewType   authViewType
	inputs     []textinput.Model
	PrevModel  types.NamedTeaModel
	focused    int
	grpcClient *grpc.GRPCClient
	err        error
	state      *types.State
}

type constructorArgs struct {
	grpcClient *grpc.GRPCClient
	viewType   authViewType
	state      *types.State
}

func NewAuthModel(args constructorArgs) *authModel {
	usernameTextInput := textinput.New()
	passwordTextInput := textinput.New()
	usernameTextInput.Placeholder = "Username"
	usernameTextInput.Focus()
	usernameTextInput.CharLimit = 32
	usernameTextInput.Width = 20

	passwordTextInput.Placeholder = "Password"
	passwordTextInput.EchoMode = textinput.EchoPassword
	passwordTextInput.EchoCharacter = '•'
	passwordTextInput.CharLimit = 32
	passwordTextInput.Width = 20

	var title string
	switch args.viewType {
	case RegisterView:
		title = "Registration"
	case AuthView:
		title = "Authorization"
	}

	return &authModel{
		inputs:     []textinput.Model{usernameTextInput, passwordTextInput},
		title:      title,
		viewType:   args.viewType,
		grpcClient: args.grpcClient,
		state:      args.state,
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
	if rm.viewType == RegisterView {
		token, err = rm.RegisterUser(username, password)
	}
	if rm.viewType == AuthView {
		token, err = rm.Authenticate(username, password)
	}

	if err != nil {
		return err
	}

	rm.state.IsAuthorized = true
	rm.state.Token = token

	return nil
}

func (rm *authModel) RegisterUser(username, password string) (string, error) {
	req := auth.RegisterRequest_builder{
		Username: proto.String(username),
		Password: proto.String(password),
	}

	resp, err := rm.grpcClient.AuthClient.Register(context.TODO(), req.Build())
	if err != nil {
		return "", err
	}

	return resp.GetToken(), nil
}

func (rm *authModel) Authenticate(username, password string) (string, error) {
	request := &auth.AuthRequest_builder{
		Username: proto.String(username),
		Password: proto.String(password),
	}

	resp, err := rm.grpcClient.AuthClient.Authenticate(context.TODO(), request.Build())
	if err != nil {
		return "", err
	}

	return resp.GetToken(), nil
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
	s += "\n\n(Press Enter to submit, Esc to go back)\n"

	return s
}
