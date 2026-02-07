package primitive

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/utils"
)

type SaveBlockFn func(ctx context.Context, b model.Block, errChan chan error)

type TextBlock struct {
	title       string
	PrevModel   types.NamedTeaModel
	Type        model.Type
	inputs      []textinput.Model
	textarea    textarea.Model
	focused     int
	isSaved     bool
	isSaving    bool
	state       *types.State
	err         error
	errChan     chan error
	saveBlockCb SaveBlockFn
	ctx         context.Context
	cancelFn    context.CancelFunc
}

type BlockArgs struct {
	Title       string
	Type        model.Type
	State       *types.State
	SaveBlockCb SaveBlockFn
}

func NewTextBlock(args BlockArgs) *TextBlock {
	passwordInput := textinput.New()
	passwordInput.Placeholder = "Enter block password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 64
	passwordInput.Width = 30

	titleInput := textinput.New()
	titleInput.Placeholder = "Enter block description"
	titleInput.CharLimit = 1024
	titleInput.Width = 50

	dataBlockInput := textarea.New()
	dataBlockInput.Placeholder = fmt.Sprintf("Enter %s data", args.Type.TypeName)
	dataBlockInput.CharLimit = 256
	dataBlockInput.Focus()

	return &TextBlock{
		Type:        args.Type,
		state:       args.State,
		inputs:      []textinput.Model{titleInput, passwordInput},
		textarea:    dataBlockInput,
		saveBlockCb: args.SaveBlockCb,
	}
}

func (r *TextBlock) GetTitle() string {
	return r.title
}

func (r *TextBlock) SetPrevModel(model types.NamedTeaModel) {
	r.PrevModel = model
}

func (r *TextBlock) Init() tea.Cmd {
	r.ctx, r.cancelFn = context.WithCancel(context.Background())
	return textinput.Blink
}

func (r *TextBlock) UpdateTextBlock(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Handle key messages
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter && msg.Alt {
			r.textarea.Blur()
			r.inputs[r.focused].Focus()
			return r, nil
		}
		switch msg.String() {
		case "esc":
			r.cancelFn()
			return r.PrevModel, nil
		case "enter":
			if !r.textarea.Focused() {
				r.focused++
			}
			if r.focused < len(r.inputs) {
				r.inputs[r.focused].Focus()
			}
			if r.focused == len(r.inputs) && r.textarea.Focused() == false && r.isSaved == false {
				title := r.inputs[0].Value()
				password := r.inputs[1].Value()
				data := r.textarea.Value()
				key, err := utils.ExtractKeyFromPassword(password, utils.ProfileMedium)
				if err != nil {
					r.err = err
					return r, nil
				}

				encrypted, nonce, err := utils.EncryptWithPassword([]byte(password), []byte(data), key)
				if err != nil {
					r.err = err

					return r, nil
				}

				r.errChan = make(chan error)
				b := model.Block{
					Title:   title,
					TypeID:  r.Type.ID,
					Profile: string(utils.ProfileMedium),
					Data:    encrypted,
					Salt:    key.Salt,
					Nonce:   nonce,
					Type:    &r.Type,
				}

				r.isSaving = true
				go r.saveBlockCb(r.ctx, b, r.errChan)

				return r, listenForSaveBlockResult(r.ctx, r.errChan)
			}
		}
	case saveBlockResultMsg:
		err := error(msg.err)
		defer close(r.errChan)
		if err != nil {
			r.err = err
			r.isSaved = false
			r.isSaving = false
			return r, nil
		}

		r.isSaving = false
		r.isSaved = true
		return r, nil
	}

	var cmd tea.Cmd
	if !r.textarea.Focused() {
		if r.focused < len(r.inputs) {
			r.inputs[r.focused], cmd = r.inputs[r.focused].Update(msg)
		}
	}

	if r.textarea.Focused() {
		r.textarea, cmd = r.textarea.Update(msg)
	}

	return r, cmd
}

func (r *TextBlock) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return r.UpdateTextBlock(msg)
}

func (r *TextBlock) TextBlockView() string {
	var s string
	if r.textarea.Focused() {
		s += r.textarea.View()
	} else {
		if r.focused < len(r.inputs) {
			s += r.inputs[r.focused].View()
		} else {
			s += "Saving block...\n"
		}
	}

	return s
}

func (r *TextBlock) View() string {
	s := fmt.Sprintf("== Creating new block: %s ==\n\n", r.Type.TypeName)
	if r.err != nil {
		s += fmt.Sprintf("Error: %s\n\n", r.err.Error())
	}
	if r.isSaving {
		s += "Saving block...\n"
		return s
	}
	if r.isSaved {
		s += "Block saved successfully! Press ESC to go back.\n"
		return s
	}

	switch r.Type.TypeName {
	case string(model.TypeNameText):
		s += r.TextBlockView()
	}

	s += "\n\nPress ESC to go back."

	return s
}
