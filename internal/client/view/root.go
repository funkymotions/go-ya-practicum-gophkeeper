package view

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
)

type rootModel struct {
	title          string
	choices        []types.NamedTeaModel // items on the to-do list
	cursor         int                   // which to-do list item our cursor is pointing at
	selected       map[int]struct{}
	MainModel      types.NamedTeaModel
	PrevModel      types.NamedTeaModel
	State          *types.State
	appChan        chan os.Signal
	storageService ports.ClientStorageService
	buildDate      string
	clientVersion  string
}

type RootModelArgs struct {
	RegisterModel  types.NamedTeaModel
	AuthModel      types.NamedTeaModel
	StorageModel   types.NamedTeaModel
	State          *types.State
	AppChan        chan os.Signal
	StorageService ports.ClientStorageService
	BuildDate      string
	ClientVersion  string
}

func NewRootModel(args RootModelArgs) *rootModel {
	mainModel := &rootModel{
		title:          "Main Menu",
		choices:        []types.NamedTeaModel{args.AuthModel, args.RegisterModel, args.StorageModel},
		selected:       make(map[int]struct{}),
		State:          args.State,
		appChan:        args.AppChan,
		storageService: args.StorageService,
		buildDate:      args.BuildDate,
		clientVersion:  args.ClientVersion,
	}

	return mainModel
}

func (cv *rootModel) Init() tea.Cmd {
	return nil
}

func (cv *rootModel) GetTitle() string {
	return cv.title
}

func (cv *rootModel) SetPrevModel(m types.NamedTeaModel) {
	cv.PrevModel = m
}

func (cv *rootModel) IsAuthorizedModel() bool {
	return false
}

func (cv *rootModel) GetState() *types.State {
	return cv.State
}

func (cv *rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			cv.storageService.UnloadToFile()
			cv.appChan <- os.Interrupt
			return cv, tea.Quit
		case "up":
			if cv.cursor > 0 {
				cv.cursor--
			}
		case "down":
			if cv.cursor < len(cv.choices)-1 {
				cv.cursor++
			}
		case "enter", " ":
			cv.choices[cv.cursor].SetPrevModel(cv)
			return cv.choices[cv.cursor], cv.choices[cv.cursor].Init()
		}
	}

	return cv, nil
}

func (cv *rootModel) View() string {
	s := fmt.Sprintf("== %s ==\n\n", cv.title)
	ready := "✅"
	notReady := "❌"
	var status string
	if cv.State.IsOnline {
		status = ready
	} else {
		status = notReady
	}

	if cv.buildDate != "" && cv.clientVersion != "" {
		s += fmt.Sprintf("Client version: %s (build date: %s)\n", cv.clientVersion, cv.buildDate)
	}

	s += "Online: " + status + "\n\n"

	for i, choice := range cv.choices {
		cursor := " " // no cursor
		if cv.cursor == i {
			cursor = ">" // cursor
		}

		checked := " " // not selected
		if _, ok := cv.selected[i]; ok {
			checked = "x" // selected
		}

		s += cursor + " [" + checked + "] " + choice.GetTitle() + "\n"
	}

	s += "\n(Press 'q' or 'Ctrl+C' to quit.)\n"

	return s
}
