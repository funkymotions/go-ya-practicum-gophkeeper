package primitive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/utils"
)

const fileMaxSize = 10 * 1024 * 1024 // 10 Megabytes

type FileBlock struct {
	Type        model.Type
	Title       string
	inputs      []textinput.Model
	isSaved     bool
	isSaving    bool
	err         error
	errChan     chan error
	focused     int
	PrevModel   types.NamedTeaModel
	saveBlockCb SaveBlockFn
	state       *types.State
	ctx         context.Context
	cancelFn    context.CancelFunc
}

func NewFileBlock(args BlockArgs) *FileBlock {
	inputDescription := textinput.New()
	inputDescription.Placeholder = "Enter block description"
	inputDescription.CharLimit = 256
	inputDescription.Width = 50
	inputDescription.Focus()

	inputFilePath := textinput.New()
	inputFilePath.Placeholder = "Enter absolute file path"
	inputFilePath.CharLimit = 256
	inputFilePath.Width = 50

	masterPasswordInput := textinput.New()
	masterPasswordInput.Placeholder = "Block Password"
	masterPasswordInput.EchoMode = textinput.EchoPassword
	masterPasswordInput.EchoCharacter = '•'
	masterPasswordInput.CharLimit = 64
	masterPasswordInput.Width = 30

	return &FileBlock{
		Type:        args.Type,
		Title:       args.Title,
		inputs:      []textinput.Model{inputDescription, inputFilePath, masterPasswordInput},
		saveBlockCb: args.SaveBlockCb,
		state:       args.State,
	}
}

func (fb *FileBlock) SetPrevModel(m types.NamedTeaModel) {
	fb.PrevModel = m
}

func (fb *FileBlock) Init() tea.Cmd {
	fb.ctx, fb.cancelFn = context.WithCancel(context.Background())

	return textinput.Blink
}

func (fb *FileBlock) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			fb.cancelFn()
			return fb.PrevModel, nil
		case "enter":
			if fb.focused == len(fb.inputs)-1 {
				title := fb.inputs[0].Value()
				filePath := fb.inputs[1].Value()
				if !filepath.IsAbs(filePath) {
					fb.err = fmt.Errorf("please provide an absolute file path")
					return fb, nil
				}

				content, err := os.ReadFile(filePath)
				if err != nil {
					fb.err = err
					return fb, nil
				}
				if len(content) > fileMaxSize {
					fb.err = fmt.Errorf("file size exceeds the maximum allowed size of %d bytes", fileMaxSize)
					return fb, nil
				}

				masterPassword := fb.inputs[2].Value()
				key, err := utils.ExtractKeyFromPassword(masterPassword, utils.ProfileMedium)
				if err != nil {
					fb.err = err
					return fb, nil
				}

				ciphertext, nonce, err := utils.EncryptWithPassword([]byte(masterPassword), content, key)
				if err != nil {
					fb.err = err
					return fb, nil
				}

				fb.errChan = make(chan error)
				b := model.Block{
					Title:   title,
					TypeID:  fb.Type.ID,
					Data:    ciphertext,
					Salt:    key.Salt,
					Nonce:   nonce,
					Profile: string(utils.ProfileMedium),
					Type:    &fb.Type,
				}

				fb.isSaving = true
				go fb.saveBlockCb(fb.ctx, b, fb.errChan)

				return fb, listenForSaveBlockResult(fb.ctx, fb.errChan)
			}

			fb.focused++
			fb.inputs[fb.focused].Focus()
		}
	case saveBlockResultMsg:
		close(fb.errChan)
		err := error(msg.err)
		if err != nil {
			fb.err = err
			fb.isSaved = false
			fb.isSaving = false
			return fb, nil
		}

		fb.isSaving = false
		fb.isSaved = true

		return fb, nil
	}

	var cmd tea.Cmd
	fb.inputs[fb.focused], cmd = fb.inputs[fb.focused].Update(msg)

	return fb, cmd
}

func (fb *FileBlock) View() string {
	if fb.isSaved {
		return "File block saved successfully!\nPress ESC to go back."
	}

	s := "Create File Block:\n"
	s += "------------------------------\n"
	if fb.err != nil {
		s += "Error: " + fb.err.Error() + "\n"
	}
	s += "\n"

	s += "Description: " + fb.inputs[0].View() + "\n"
	s += "File Path: " + fb.inputs[1].View() + "\n"
	s += "Block Password: " + fb.inputs[2].View() + "\n"

	s += "\nPress Enter to save the block or ESC to go back.\n"

	return s
}
