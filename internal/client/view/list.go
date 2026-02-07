package view

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
)

type blockListView struct {
	title             string
	PrevModel         types.NamedTeaModel
	cursor            int
	selected          map[int]struct{}
	service           ports.ClientStorageService
	isAuthorizedModel bool
	state             *types.State
	blocks            *[]*model.Block
	err               error
	updateCh          chan struct{}
	errUpdateCh       chan error
}

type BlockListViewArgs struct {
	State   *types.State
	Service ports.ClientStorageService
	// BlocksChan chan []*model.Block
	// ErrChan chan error
	UpdateCh    chan struct{}
	Blocks      *[]*model.Block
	ErrUpdateCh chan error
}

func NewBlockListView(args BlockListViewArgs) *blockListView {
	s := &blockListView{
		title:       "Storage",
		selected:    make(map[int]struct{}),
		service:     args.Service,
		state:       args.State,
		blocks:      args.Blocks,
		updateCh:    args.UpdateCh,
		errUpdateCh: args.ErrUpdateCh,
	}

	return s
}

func (sm *blockListView) IsAuthorizedModel() bool {
	return true
}

type MsgBlocksReceived struct{}

func (sm *blockListView) Init() tea.Cmd {
	return sm.listenForBlocks()
}

func (sm *blockListView) listenForBlocks() tea.Cmd {
	return func() tea.Msg {
		select {
		case err := <-sm.errUpdateCh:
			return err
		case <-sm.updateCh:
			return MsgBlocksReceived{}
		}
	}
}

func (sm *blockListView) GetTitle() string {
	return sm.title
}

func (sm *blockListView) SetPrevModel(m types.NamedTeaModel) {
	sm.PrevModel = m
}

func (sm *blockListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return sm.PrevModel, nil
		case "down":
			if sm.cursor < len(*sm.blocks)-1 {
				sm.cursor++
			}
		case "up":
			if sm.cursor > 0 {
				sm.cursor--
			}
		case "enter":
			if len(*sm.blocks) == 0 {
				return sm, nil
			}

			block := (*sm.blocks)[sm.cursor]
			blockModel := NewBlockModel(sm)
			blockModel.block = *block

			return blockModel, blockModel.Init()
		}

		return sm, nil
	case MsgBlocksReceived:
		return sm, sm.listenForBlocks()
	case error:
		sm.err = msg
		return sm, sm.listenForBlocks()
	}

	return sm, nil
}

func (sm *blockListView) View() string {
	s := "== Storage ==\n\n"
	var errText string
	if sm.err != nil {
		errText = "Error: " + sm.err.Error() + "\n\n"
	}
	s += errText

	if len(*sm.blocks) == 0 {
		s += "No data blocks available.\n"
	} else {
		s += "Data Blocks:\n"
		for i, block := range *sm.blocks {
			cursor := " "
			if sm.cursor == i {
				cursor = ">"
			}

			s += fmt.Sprintf("%s %4d) %-4d %-40s %-20s\n", cursor, i+1, block.ID, block.Title, block.Type.TypeName)
		}
	}

	s += "\n(Press 'ESC' to go back.)\n"

	return s
}
