package view

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/primitive"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
)

type addBlockView struct {
	PrevModel types.NamedTeaModel
	State     *types.State
	title     string
	service   ports.ClientStorageService
	types     []*model.Type
	cursor    int
	err       error
	typeChan  chan []*model.Type
	errChan   chan error
	ctx       context.Context
	cancelFn  context.CancelFunc
}

type AddBlockViewArgs struct {
	State   *types.State
	Service ports.ClientStorageService
}

func NewAddBlockView(args AddBlockViewArgs) *addBlockView {
	return &addBlockView{
		title:    "Add Block",
		State:    args.State,
		service:  args.Service,
		typeChan: make(chan []*model.Type, 1),
		errChan:  make(chan error, 1),
	}
}

func (abv *addBlockView) GetTitle() string {
	return abv.title
}

func (abv *addBlockView) SetPrevModel(m types.NamedTeaModel) {
	abv.PrevModel = m
}

func (abv *addBlockView) IsAuthorizedModel() bool {
	return true
}

func (abv *addBlockView) Init() tea.Cmd {
	abv.typeChan = make(chan []*model.Type, 1)
	abv.errChan = make(chan error, 1)
	abv.ctx, abv.cancelFn = context.WithCancel(context.Background())
	go abv.service.ListBlockTypes(abv.ctx, abv.typeChan, abv.errChan)

	return abv.listenForTypes()
}

type MsgBlockTypeReceived []*model.Type

func (abv *addBlockView) listenForTypes() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-abv.ctx.Done():
			return nil
		case err := <-abv.errChan:
			abv.err = err
			return nil
		case types := <-abv.typeChan:
			return MsgBlockTypeReceived(types)
		}
	}
}

func (abv *addBlockView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			abv.cancelFn()
			close(abv.typeChan)
			close(abv.errChan)
			return abv.PrevModel, nil
		case "down":
			if abv.cursor < len(abv.types)-1 {
				abv.cursor++
			}
		case "up":
			if abv.cursor > 0 {
				abv.cursor--
			}
		case "enter":
			if len(abv.types) == 0 {
				abv.err = fmt.Errorf("authorization required to add block")
				return abv, nil
			}
			selectedType := abv.types[abv.cursor]
			switch selectedType.TypeName {
			case string(model.TypeNameText):
				// Raw text
				m := primitive.NewTextBlock(primitive.BlockArgs{
					State:       abv.State,
					Type:        *selectedType,
					SaveBlockCb: abv.service.SaveBlock,
				})
				m.SetPrevModel(abv)
				return m, m.Init()
			case string(model.TypeNameCredentials):
				// Credentials
				m := primitive.NewCredentialsBlock(
					primitive.BlockArgs{
						State:       abv.State,
						Type:        *selectedType,
						SaveBlockCb: abv.service.SaveBlock,
					},
				)
				m.SetPrevModel(abv)
				return m, m.Init()
			case string(model.TypeNameCard):
				// Bank Card
				m := primitive.NewBankCardBlock(
					primitive.BlockArgs{
						State:       abv.State,
						Type:        *selectedType,
						SaveBlockCb: abv.service.SaveBlock,
					},
				)
				m.SetPrevModel(abv)
				return m, m.Init()
			case string(model.TypeNameFile):
				// File
				m := primitive.NewFileBlock(
					primitive.BlockArgs{
						State:       abv.State,
						Type:        *selectedType,
						SaveBlockCb: abv.service.SaveBlock,
					},
				)
				m.SetPrevModel(abv)
				return m, m.Init()
			}

			return abv, nil
		}
	case MsgBlockTypeReceived:
		abv.types = msg
		return abv, nil
	}

	return abv, nil
}

func (abv *addBlockView) View() string {
	s := "== Add New Block ==\n\n"
	if len(abv.types) != 0 {
		cursor := abv.cursor
		for i, t := range abv.types {
			if i == cursor {
				s += "> "
			} else {
				s += "  "
			}
			s += t.TypeName + "\n"
		}
	}
	s += "(Press 'ESC' to go back.)\n"

	return s
}
