package client

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
)

type storageModel struct {
	PrevModel types.NamedTeaModel
	State     *types.State
	choices   []string
	selected  map[int]struct{}
	cursor    int
	title     string
}

func NewStorageModel(state *types.State) *storageModel {
	s := &storageModel{
		title:    "Storage Menu",
		selected: make(map[int]struct{}),
		State:    state,
		choices:  []string{"View Blocks", "Add Block"},
	}

	return s
}

func (sm *storageModel) GetTitle() string {
	if sm.State.IsAuthorized {
		return sm.title
	} else {
		return "==Auth required== " + sm.title
	}
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
			case 0:
				blockListView := NewBlockListView(sm.State)
				blockListView.SetPrevModel(sm)

				return blockListView, blockListView.Init()
			case 1:
				addBlockView := NewAddBlockView(sm.State)
				addBlockView.SetPrevModel(sm)

				return addBlockView, addBlockView.Init()
			}
		}
	}

	return sm, nil
}

func (sm *storageModel) View() string {
	s := "Storage Menu:\n\n"
	for i, choice := range sm.choices {
		cursor := " " // no cursor
		if sm.cursor == i {
			cursor = ">" // cursor
		}

		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}
