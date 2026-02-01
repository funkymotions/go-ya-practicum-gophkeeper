package client

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	grpcclient "github.com/funkymotions/go-ya-practicum-gophkeeper/internal/infrastructure/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

type modelView struct {
	title     string
	choices   []types.NamedTeaModel // items on the to-do list
	cursor    int                   // which to-do list item our cursor is pointing at
	selected  map[int]struct{}
	MainModel types.NamedTeaModel
	PrevModel types.NamedTeaModel
	State     *types.State
	g         *grpc.ClientConn
}

func New() *modelView {
	g, err := grpc.NewClient(
		"127.0.0.1:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  time.Second,
				Multiplier: 1,
				MaxDelay:   time.Second,
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	go func() {
		g.WaitForStateChange(context.Background(), g.GetState())
	}()

	state := types.NewState()

	// defer g.Close()
	client := grpcclient.NewGRPCClient(g)
	registerModel := NewAuthModel(
		constructorArgs{
			grpcClient: client,
			viewType:   RegisterView,
			state:      state,
		},
	)

	authModel := NewAuthModel(
		constructorArgs{
			grpcClient: client,
			viewType:   AuthView,
			state:      state,
		},
	)

	storageModel := NewStorageModel(state)

	mainModel := &modelView{
		title:    "Main Menu",
		choices:  []types.NamedTeaModel{authModel, registerModel, storageModel},
		selected: make(map[int]struct{}),
		State:    state,
		g:        g,
	}

	return mainModel
}

func (cv *modelView) Init() tea.Cmd {
	return nil
}

func (cv *modelView) GetTitle() string {
	return cv.title
}

func (cv *modelView) SetPrevModel(m types.NamedTeaModel) {
	cv.PrevModel = m
}

func (cv *modelView) IsAuthorizedModel() bool {
	return false
}

func (cv *modelView) GetState() *types.State {
	return cv.State
}

func (cv *modelView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return cv, tea.Quit
		case "up", "k":
			if cv.cursor > 0 {
				// if cv.State.IsAuthorized != cv.choices[cv.cursor-1].IsAuthorizedModel() {
				// break
				// }
				cv.cursor--
			}
		case "down", "j":
			if cv.cursor < len(cv.choices)-1 {
				// if cv.State.IsAuthorized != cv.choices[cv.cursor+1].IsAuthorizedModel() {
				// break
				// }
				cv.cursor++
			}
		case "enter", " ":
			cv.choices[cv.cursor].SetPrevModel(cv)

			return cv.choices[cv.cursor], cv.choices[cv.cursor].Init()
		}
	}

	return cv, nil
}

func (cv *modelView) View() string {
	s := fmt.Sprintf("== %s ==\n\n", cv.title)
	connReady := "✅"
	connNotReady := "❌"
	connState := cv.g.GetState().String()
	status := ""
	if connState != "READY" {
		status = connNotReady
	} else {
		status = connReady
	}

	s += "GRPC State: " + status + " IsAuthorized: " + fmt.Sprintf("%v", cv.State.IsAuthorized) + "\n\n"

	for i, choice := range cv.choices {
		cursor := " " // no cursor
		if cv.cursor == i {
			cursor = ">" // cursor
		}

		checked := " " // not selected
		if _, ok := cv.selected[i]; ok {
			checked = "x" // selected
		}

		// if cv.State.IsAuthorized != choice.IsAuthorizedModel() {
		// s += ""
		// } else {
		s += cursor + " [" + checked + "] " + choice.GetTitle() + "\n"
		// }
	}

	s += "\n[!]Press q to quit.\n"

	return s
}
