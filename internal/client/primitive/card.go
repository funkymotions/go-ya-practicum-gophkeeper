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

type BankCardBlock struct {
	Title       string
	PrevModel   types.NamedTeaModel
	Type        model.Type
	CardNumber  string
	ExpiryDate  string
	CardHolder  string
	CVV         string
	State       *types.State
	inputs      []textinput.Model
	focused     int
	isSaved     bool
	isSaving    bool
	err         error
	errChan     chan error
	saveBlockCb SaveBlockFn
	ctx         context.Context
	cancelFn    context.CancelFunc
}

func NewBankCardBlock(args BlockArgs) *BankCardBlock {
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter block description"
	titleInput.CharLimit = 256
	titleInput.Width = 50
	titleInput.Focus()

	cardNumberInput := textinput.New()
	cardNumberInput.Placeholder = "Card Number"
	cardNumberInput.CharLimit = 16
	cardNumberInput.Width = 20

	expiryDateInput := textinput.New()
	expiryDateInput.Placeholder = "Expiry Date (MM/YY)"
	expiryDateInput.CharLimit = 5
	expiryDateInput.Width = 10

	cardHolderInput := textinput.New()
	cardHolderInput.Placeholder = "Card Holder Name"
	cardHolderInput.CharLimit = 64
	cardHolderInput.Width = 30

	cvvInput := textinput.New()
	cvvInput.Placeholder = "CVV"
	cvvInput.CharLimit = 4
	cvvInput.Width = 5

	masterPasswordInput := textinput.New()
	masterPasswordInput.Placeholder = "Master Password"
	masterPasswordInput.EchoMode = textinput.EchoPassword
	masterPasswordInput.EchoCharacter = '•'
	masterPasswordInput.CharLimit = 64
	masterPasswordInput.Width = 30

	return &BankCardBlock{
		Type:  args.Type,
		State: args.State,
		inputs: []textinput.Model{
			titleInput,
			cardNumberInput,
			expiryDateInput,
			cardHolderInput,
			cvvInput,
			masterPasswordInput,
		},
		saveBlockCb: args.SaveBlockCb,
	}
}

func (bcb *BankCardBlock) Init() tea.Cmd {
	bcb.ctx, bcb.cancelFn = context.WithCancel(context.Background())

	return textinput.Blink
}

func (bcb *BankCardBlock) GetTitle() string {
	return bcb.Title
}

func (bcb *BankCardBlock) SetPrevModel(model types.NamedTeaModel) {
	bcb.PrevModel = model
}

func (bcb *BankCardBlock) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			bcb.cancelFn()
			return bcb.PrevModel, nil
		case "enter":
			if bcb.focused == len(bcb.inputs)-1 {
				bcb.Title = bcb.inputs[0].Value()
				bcb.CardNumber = bcb.inputs[1].Value()
				bcb.ExpiryDate = bcb.inputs[2].Value()
				bcb.CardHolder = bcb.inputs[3].Value()
				bcb.CVV = bcb.inputs[4].Value()
				masterPassword := bcb.inputs[5].Value()
				key, err := utils.ExtractKeyFromPassword(masterPassword, utils.ProfileMedium)
				if err != nil {
					bcb.err = err

					return bcb, nil
				}

				payload := fmt.Sprintf(
					"Title: %s\nPAN: %s\nExpiry: %s\nCard holder: %s\nCVV: %s\n",
					bcb.Title,
					bcb.CardNumber,
					bcb.ExpiryDate,
					bcb.CardHolder,
					bcb.CVV,
				)

				ciphertext, nonce, err := utils.EncryptWithPassword([]byte(masterPassword), []byte(payload), key)
				if err != nil {
					bcb.err = err

					return bcb, nil
				}

				bcb.errChan = make(chan error)
				b := model.Block{
					Title:   bcb.Title,
					TypeID:  bcb.Type.ID,
					Data:    ciphertext,
					Salt:    key.Salt,
					Nonce:   nonce,
					Profile: string(utils.ProfileMedium),
					Type:    &bcb.Type,
				}

				bcb.isSaving = true
				go bcb.saveBlockCb(bcb.ctx, b, bcb.errChan)

				return bcb, listenForSaveBlockResult(bcb.ctx, bcb.errChan)
			}

			bcb.focused++
			bcb.inputs[bcb.focused].Focus()
		}
	case saveBlockResultMsg:
		close(bcb.errChan)
		err := error(msg.err)
		if err != nil {
			bcb.err = err
			bcb.isSaved = false
			bcb.isSaving = false

			return bcb, nil
		}
		bcb.isSaved = true
		bcb.isSaving = false

		return bcb, nil
	}

	var cmd tea.Cmd
	bcb.inputs[bcb.focused], cmd = bcb.inputs[bcb.focused].Update(msg)

	return bcb, cmd
}

func (bcb *BankCardBlock) View() string {
	if bcb.isSaving {
		return "Saving block...\n"
	}
	if bcb.isSaved {
		return "Bank Card block saved successfully!\n\nPress 'ESC' key to return."
	}
	s := "== Add New Bank Card Block ==\n\n"
	for i, input := range bcb.inputs {
		cursor := " " // no cursor
		if bcb.focused == i {
			cursor = ">" // cursor
		}
		s += cursor + " " + input.View() + "\n"
	}
	if bcb.err != nil {
		s += "\nError: " + bcb.err.Error() + "\n"
	}
	s += "\nPress 'Enter' to save the block or 'ESC' to cancel.\n"

	return s
}
