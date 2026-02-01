package client

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/infrastructure/grpc"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/storage"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/subscription"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type blockListView struct {
	title             string
	PrevModel         types.NamedTeaModel
	cursor            int
	selected          map[int]struct{}
	grpcClient        *grpc.GRPCClient
	isAuthorizedModel bool
	state             *types.State
	blocks            []*model.Block
	blocksChan        chan []*model.Block
	err               error
}

func NewBlockListView(state *types.State) *blockListView {
	s := &blockListView{
		title:      "Storage",
		selected:   make(map[int]struct{}),
		grpcClient: grpc.NewGRPCClient(),
		state:      state,
		blocks:     []*model.Block{},
		blocksChan: make(chan []*model.Block),
	}

	return s
}

func (sm *blockListView) IsAuthorizedModel() bool {
	return true
}

type MsgBlocksReceived []*model.Block

func (sm *blockListView) Init() tea.Cmd {
	sm.blocks = make([]*model.Block, 0)
	go sm.startBlockStream()

	return sm.listenForBlocks()
}

func (sm *blockListView) listenForBlocks() tea.Cmd {
	return func() tea.Msg {
		blocks, ok := <-sm.blocksChan
		if !ok {
			return nil
		}
		return MsgBlocksReceived(blocks)
	}
}

func (sm *blockListView) GetTitle() string {
	if sm.state.IsAuthorized {
		return sm.title
	} else {
		return "==Auth required== " + sm.title
	}
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
			if sm.cursor < len(sm.blocks)-1 {
				sm.cursor++
			}
		case "up":
			if sm.cursor > 0 {
				sm.cursor--
			}
		case "enter":
			if len(sm.blocks) == 0 {
				return sm, nil
			}
			block := sm.blocks[sm.cursor]
			blockModel := NewBlockModel(sm)
			blockModel.block = *block

			return blockModel, blockModel.Init()
		}
	case MsgBlocksReceived:
		sm.blocks = msg

		return sm, sm.listenForBlocks()
	}

	return sm, nil
}

func (sm *blockListView) startBlockStream() {
	req := storage.ListDataBlocksRequest_builder{
		ClientId: proto.String(sm.state.ClientID),
	}

	md := metadata.New(map[string]string{
		"authorization": sm.state.Token,
	})

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	sub := subscription.SubscribeRequest_builder{
		ClientId: &sm.state.ClientID,
	}.Build()

	_, err := sm.grpcClient.SubscriptionClient.Subscribe(ctx, sub)
	if err != nil {
		sm.err = err
		return
	}

	stream, err := sm.grpcClient.StorageClient.ListDataBlocks(ctx, req.Build())
	if err != nil {
		sm.err = err
		return
	}

	go func() {
		for {
			blockResp, err := stream.Recv()
			if err != nil {
				sm.err = err
				close(sm.blocksChan)
				return
			}

			listBlocks := blockResp.GetDataBlocks()
			blocks := make([]*model.Block, 0, len(listBlocks))
			for _, blockResp := range listBlocks {
				t := blockResp.GetType()
				block := &model.Block{
					ID:      int(blockResp.GetBlockId()),
					Title:   blockResp.GetTitle(),
					Data:    blockResp.GetChiphertext(),
					Salt:    blockResp.GetSalt(),
					Nonce:   blockResp.GetNonce(),
					Profile: blockResp.GetProfile().String(),
					Type: &model.Type{
						ID:          int(t.GetId()),
						TypeName:    t.GetTypeName(),
						Description: t.GetDescription(),
					},
				}

				blocks = append(blocks, block)
			}

			sm.blocksChan <- blocks
		}
	}()
}

func (sm *blockListView) View() string {
	s := "== Storage ==\n\n"
	var errText string
	if sm.err != nil {
		errText = "Error: " + sm.err.Error() + "\n\n"
	}
	s += errText

	if len(sm.blocks) == 0 {
		s += "No data blocks available.\n"
	} else {
		s += "Data Blocks:\n"
		for i, block := range sm.blocks {
			cursor := " " // no cursor
			if sm.cursor == i {
				cursor = ">" // cursor!
			}

			s += fmt.Sprintf("%s %-30s %-20s\n", cursor, block.Title, block.Type.TypeName)
		}
	}

	s += "\nPress 'esc' to return to the main menu.\n"

	return s
}
