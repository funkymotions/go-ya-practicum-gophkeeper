package view

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
)

type storageModel struct {
	PrevModel     types.NamedTeaModel
	state         *types.State
	choices       []string
	selected      map[int]struct{}
	cursor        int
	title         string
	addBlockView  types.NamedTeaModel
	listBlockView types.NamedTeaModel
}

type StorageModelArgs struct {
	State          *types.State
	AddBlockModel  types.NamedTeaModel
	ListBlockModel types.NamedTeaModel
}

func NewStorageModel(args StorageModelArgs) *storageModel {
	s := &storageModel{
		title:         "Storage Menu",
		selected:      make(map[int]struct{}),
		state:         args.State,
		choices:       []string{"View Blocks", "Add Block"},
		addBlockView:  args.AddBlockModel,
		listBlockView: args.ListBlockModel,
	}

	return s
}

func (sm *storageModel) GetTitle() string {
	return sm.title
}

func (sm *storageModel) SetPrevModel(m types.NamedTeaModel) {
	sm.PrevModel = m
}

func (sm *storageModel) IsAuthorizedModel() bool {
	return true
}

func (sm *storageModel) Init() tea.Cmd {
	return nil
}

func (sm *storageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return sm.PrevModel, nil
		case "up":
			if sm.cursor > 0 {
				sm.cursor--
			}
		case "down":
			if sm.cursor < len(sm.choices)-1 {
				sm.cursor++
			}
		case "enter", " ":
			switch sm.cursor {
			case 1:
				sm.addBlockView.SetPrevModel(sm)
				return sm.addBlockView, sm.addBlockView.Init()
			case 0:
				sm.listBlockView.SetPrevModel(sm)
				return sm.listBlockView, sm.listBlockView.Init()
			}
		}
	}

	return sm, nil
}

func (sm *storageModel) View() string {
	s := "== Storage Menu ==\n\n"
	for i, choice := range sm.choices {
		cursor := " " // no cursor
		if sm.cursor == i {
			cursor = ">" // cursor
		}

		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\n(Press 'ESC' to go back.)\n"

	return s
}
