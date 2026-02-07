package primitive

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/utils"
)

type credentialsBlock struct {
	Type        model.Type
	Username    string
	Password    string
	Title       string
	State       *types.State
	inputs      []textinput.Model
	focused     int
	PrevModel   types.NamedTeaModel
	saveBlockCb SaveBlockFn
	err         error
	errChan     chan error
	isSaved     bool
	isSaving    bool
	ctx         context.Context
	cancelFn    context.CancelFunc
}

func NewCredentialsBlock(args BlockArgs) *credentialsBlock {
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter block description"
	titleInput.CharLimit = 256
	titleInput.Width = 50
	titleInput.Focus()

	usernameInput := textinput.New()
	usernameInput.Placeholder = "Username"
	usernameInput.CharLimit = 32
	usernameInput.Width = 20

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 32
	passwordInput.Width = 20

	masterPasswordInput := textinput.New()
	masterPasswordInput.Placeholder = "Master Password"
	masterPasswordInput.EchoMode = textinput.EchoPassword
	masterPasswordInput.EchoCharacter = '•'
	masterPasswordInput.CharLimit = 64
	masterPasswordInput.Width = 30

	return &credentialsBlock{
		inputs:      []textinput.Model{titleInput, usernameInput, passwordInput, masterPasswordInput},
		Type:        args.Type,
		State:       args.State,
		saveBlockCb: args.SaveBlockCb,
	}
}

func (c *credentialsBlock) Init() tea.Cmd {
	c.ctx, c.cancelFn = context.WithCancel(context.Background())

	return textinput.Blink
}

func (c *credentialsBlock) SetPrevModel(m types.NamedTeaModel) {
	c.PrevModel = m
}

func (c *credentialsBlock) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			c.cancelFn()
			return c.PrevModel, nil
		case "enter":
			if c.focused == len(c.inputs)-1 {
				c.Title = c.inputs[0].Value()
				c.Username = c.inputs[1].Value()
				c.Password = c.inputs[2].Value()
				masterPassword := c.inputs[3].Value()
				key, err := utils.ExtractKeyFromPassword(masterPassword, utils.ProfileMedium)
				if err != nil {
					c.err = err
					return c, nil
				}

				payload := fmt.Sprintf("username:%s / password:%s", c.Username, c.Password)
				encrypted, nonce, err := utils.EncryptWithPassword([]byte(masterPassword), []byte(payload), key)
				if err != nil {
					c.err = err
					return c, nil
				}

				c.errChan = make(chan error)
				b := model.Block{
					Title:   c.Title,
					TypeID:  c.Type.ID,
					Data:    encrypted,
					Salt:    key.Salt,
					Nonce:   nonce,
					Profile: string(utils.ProfileMedium),
					Type:    &c.Type,
				}

				c.isSaving = true
				go c.saveBlockCb(c.ctx, b, c.errChan)

				return c, listenForSaveBlockResult(c.ctx, c.errChan)
			}

			c.focused++
			c.inputs[c.focused-1].Blur()
			c.inputs[c.focused].Focus()
		}
	case saveBlockResultMsg:
		close(c.errChan)
		err := error(msg.err)
		if err != nil {
			c.err = err
			c.isSaved = false
			c.isSaving = false

			return c, nil
		}
		c.isSaving = false
		c.isSaved = true

		return c, nil
	}

	var cmd tea.Cmd
	c.inputs[c.focused], cmd = c.inputs[c.focused].Update(msg)

	return c, cmd
}

func (c *credentialsBlock) View() string {
	if c.isSaved {
		return "Credentials block saved successfully! Press 'ESC' to go back.\n"
	}

	if c.isSaving {
		return "Saving block...\n"
	}

	s := "Enter your credentials:\n\n"
	var errText string
	if c.err != nil {
		errText = fmt.Sprintf("Error: %v\n\n", c.err)
	}
	s += errText

	for i, input := range c.inputs {
		if i == c.focused {
			s += "* " + input.View() + "\n"
		} else {
			s += "* " + input.View() + "\n"
		}
	}

	s += "\nPress 'ESC' to go back. Press Enter to submit inputs.\n"

	return s
}
