package view

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/utils"
)

type blockModel struct {
	prevModel     types.NamedTeaModel
	block         model.Block
	passInput     textinput.Model
	focused       int
	decryptedText []byte
	err           error
}

func NewBlockModel(prevModel types.NamedTeaModel) *blockModel {
	passwordInput := textinput.New()
	passwordInput.Placeholder = "Enter block password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 64
	passwordInput.Width = 30
	passwordInput.Focus()

	return &blockModel{
		prevModel: prevModel,
		passInput: passwordInput,
	}
}

func (bm *blockModel) SaveFileBlockToDisk(blockName string, content []byte) error {
	mimeType := http.DetectContentType(content[:512])
	fileExt, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(fileExt) == 0 {
		return fmt.Errorf("Could not determine file extension\n\n\n")
	}

	cwd, err := os.Getwd()
	if err != nil {
		bm.err = fmt.Errorf("Error getting current working directory: %v", err)
		return err
	}

	filePath := filepath.Join(cwd, blockName+fileExt[0])
	err = os.WriteFile(filePath, content, 0644)
	if err != nil {
		bm.err = fmt.Errorf("Error writing file to disk: %v", err)
		return err
	}

	return nil
}

func (bm *blockModel) Init() tea.Cmd {
	return textinput.Blink
}

func (bm *blockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return bm.prevModel, nil
		case "enter":
			bm.err = nil
			password := bm.passInput.Value()
			decrypted, err := utils.DecryptWithPassword(
				bm.block.Data,
				bm.block.Nonce,
				[]byte(password),
				bm.block.Salt,
				utils.ScryptProfile(bm.block.Profile),
			)
			if err != nil {
				bm.err = fmt.Errorf("Invalid password")
				bm.passInput.Reset()
				return bm, nil
			}

			bm.decryptedText = decrypted
			if bm.block.Type.TypeName == string(model.TypeNameFile) {
				err = bm.SaveFileBlockToDisk(bm.block.Title, decrypted)
				if err != nil {
					bm.err = err
					return bm, nil
				}
			}

			return bm, nil
		}
	}

	var cmd tea.Cmd
	bm.passInput, cmd = bm.passInput.Update(msg)

	return bm, cmd
}

func (bm *blockModel) View() string {
	var errText string
	if bm.err != nil {
		errText = fmt.Sprintf("\nError: %s", bm.err.Error())
	}

	if bm.block.Type.TypeName == string(model.TypeNameFile) && len(bm.decryptedText) != 0 {
		s := "File block has been decrypted and saved to disk.\n"
		s += errText
		s += "\n\nPress 'esc' to go back.\n"
		return s
	}

	s := "Block details:\n"
	s += strings.Repeat("_", 40) + "\n"
	s += errText + "\n"
	s += fmt.Sprintf("ID: %d | Type: %s | Title: %s\n", bm.block.ID, bm.block.Type.TypeName, bm.block.Title)
	s += strings.Repeat("_", 40) + "\n\n"
	if len(bm.decryptedText) > 0 {
		del := strings.Repeat("_", 40) + "\n"
		s += fmt.Sprintf("Block data:\n%s\n%s\n", del, string(bm.decryptedText))
		s += strings.Repeat("_", 40) + "\n"
	} else {
		s += bm.passInput.View()
	}

	s += "\n\n(Press 'ESC' to go back.)\n"

	return s
}
