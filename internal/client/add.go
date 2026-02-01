package client

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/blocks"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	grpcclient "github.com/funkymotions/go-ya-practicum-gophkeeper/internal/infrastructure/grpc"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/storage"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/utils"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type addBlockView struct {
	PrevModel types.NamedTeaModel
	State     *types.State
	title     string
	client    *grpcclient.GRPCClient
	types     []*model.Type
	cursor    int
	typeChan  chan []*model.Type
}

func NewAddBlockView(state *types.State) *addBlockView {
	return &addBlockView{
		title:    "Add Block",
		State:    state,
		client:   grpcclient.NewGRPCClient(),
		typeChan: make(chan []*model.Type),
	}
}

func (abv *addBlockView) GetTitle() string {
	if abv.State.IsAuthorized {
		return abv.title
	} else {
		return "==Auth required== " + abv.title
	}
}

func (abv *addBlockView) SetPrevModel(m types.NamedTeaModel) {
	abv.PrevModel = m
}

func (abv *addBlockView) IsAuthorizedModel() bool {
	return true
}

func (abv *addBlockView) Init() tea.Cmd {
	go abv.ListBlockTypes()

	return abv.listenForTypes()
}

type MsgBlockTypeReceived []*model.Type

func (abv *addBlockView) listenForTypes() tea.Cmd {
	return func() tea.Msg {
		types, ok := <-abv.typeChan
		if !ok {
			return nil
		}
		return MsgBlockTypeReceived(types)
	}
}

func (r *addBlockView) SaveBlock(
	title string,
	typeID int,
	profile utils.ScryptProfile,
	data []byte,
	salt []byte,
	nonce []byte,
) error {
	md := metadata.Pairs("authorization", r.State.Token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	req := storage.SaveDataBlockRequest_builder{
		Title:       proto.String(title),
		TypeId:      proto.Int32(int32(typeID)),
		Chiphertext: data,
		Salt:        salt,
		Nonce:       nonce,
		Profile:     storage.EncProfile(utils.ProfileToProto(profile)).Enum(),
	}

	_, err := r.client.StorageClient.SaveDataBlock(ctx, req.Build())
	if err != nil {
		return err
	}

	return nil
}

func (abv *addBlockView) ListBlockTypes() error {
	md := metadata.New(map[string]string{
		"authorization": abv.State.Token,
	})

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	req := storage.GetBlockTypesRequest_builder{}.Build()
	resp, err := abv.client.StorageClient.ListBlockTypes(ctx, req)
	if err != nil {
		return err
	}

	blockType := resp.GetBlockTypes()
	types := make([]*model.Type, len(blockType))
	for i, bt := range blockType {
		types[i] = &model.Type{
			ID:          int(bt.GetId()),
			TypeName:    bt.GetTypeName(),
			Description: bt.GetDescription(),
		}
	}

	abv.typeChan <- types

	return nil
}

func (abv *addBlockView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
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
			selectedType := abv.types[abv.cursor]
			switch selectedType.TypeName {
			case string(model.TypeNameText):
				// Raw text
				m := blocks.NewTextBlock(blocks.BlockArgs{
					State:       abv.State,
					Type:        *selectedType,
					SaveBlockCb: abv.SaveBlock,
				})

				m.SetPrevModel(abv)

				return m, m.Init()

			case string(model.TypeNameCredentials):
				// Credentials
				m := blocks.NewCredentialsBlock(
					blocks.BlockArgs{
						State:       abv.State,
						Type:        *selectedType,
						SaveBlockCb: abv.SaveBlock,
					},
				)
				m.SetPrevModel(abv)

				return m, m.Init()
			case string(model.TypeNameCard):
				// Bank Card
				m := blocks.NewBankCardBlock(
					blocks.BlockArgs{
						State:       abv.State,
						Type:        *selectedType,
						SaveBlockCb: abv.SaveBlock,
					},
				)
				m.SetPrevModel(abv)

				return m, m.Init()

			case string(model.TypeNameFile):
				// File
				m := blocks.NewFileBlock(
					blocks.BlockArgs{
						State:       abv.State,
						Type:        *selectedType,
						SaveBlockCb: abv.SaveBlock,
					},
				)
				m.SetPrevModel(abv)

				return m, m.Init()
			}

			return abv, nil
		}
	case MsgBlockTypeReceived:
		abv.types = msg
		return abv, abv.listenForTypes()
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
	s += "Press 'esc' to go back.\n"

	return s
}
