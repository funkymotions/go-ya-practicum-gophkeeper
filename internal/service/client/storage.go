package client

import (
	"context"
	"fmt"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/storage"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/utils"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type offlineRepo interface {
	ports.OfflineFileHandler
	ports.OfflineBlockReader
	ports.OfflineBlockWriter
	ports.OfflineTypesReader
	ports.OfflineTypesWriter
	ports.OfflineAppStateReader
}

type clientStorageService struct {
	client            storage.StorageServiceClient
	offlineRepository offlineRepo
	state             *types.State
}

type ClientStorageServiceArgs struct {
	Client            storage.StorageServiceClient
	State             *types.State
	OfflineRepository offlineRepo
}

var _ ports.ClientStorageService = (*clientStorageService)(nil)

func NewClientStorageService(args ClientStorageServiceArgs) *clientStorageService {
	return &clientStorageService{
		client:            args.Client,
		state:             args.State,
		offlineRepository: args.OfflineRepository,
	}
}

func (s *clientStorageService) SaveBlock(
	ctx context.Context,
	b model.Block,
	errChan chan error,
) {
	if !s.state.IsOnline {
		s.offlineRepository.SaveUnsyncedBlock(b)
		select {
		case <-ctx.Done():
			return
		case errChan <- nil:
			return
		}
	}

	md := metadata.Pairs("authorization", s.state.Token)
	ctx = metadata.NewOutgoingContext(ctx, md)
	profile := utils.ScryptProfile(b.Profile)
	req := storage.SaveDataBlockRequest_builder{
		Title:       proto.String(b.Title),
		TypeId:      proto.Int32(int32(b.TypeID)),
		Chiphertext: b.Data,
		Salt:        b.Salt,
		Nonce:       b.Nonce,
		Profile:     storage.EncProfile(utils.ProfileToProto(profile)).Enum(),
	}

	_, err := s.client.SaveDataBlock(ctx, req.Build())
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChan <- err:
			return
		}
	}

	select {
	case <-ctx.Done():
		return
	case errChan <- nil:
		s.offlineRepository.SaveSyncedBlock(b)
		return
	}
}

func (s *clientStorageService) ListBlockTypes(
	ctx context.Context,
	typesChan chan []*model.Type,
	errChan chan error,
) {
	if !s.state.IsOnline {
		types := s.offlineRepository.ReadTypes()
		select {
		case <-ctx.Done():
			return
		case typesChan <- types:
			return
		}
	}

	md := metadata.New(map[string]string{
		"authorization": s.state.Token,
	})

	ctx = metadata.NewOutgoingContext(ctx, md)
	req := storage.GetBlockTypesRequest_builder{}.Build()
	resp, err := s.client.ListBlockTypes(ctx, req)
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChan <- err:
			return
		}
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

	select {
	case <-ctx.Done():
		return
	case typesChan <- types:
		s.offlineRepository.SaveTypes(types)
		return
	}
}

func (s *clientStorageService) StartBlockStream(
	ctx context.Context,
	blocksChan chan []*model.Block,
	errChan chan error,
) {
	if !s.state.IsOnline {
		blocks := s.offlineRepository.ReadBlocks()
		select {
		case <-ctx.Done():
			return
		case blocksChan <- blocks:
			return
		default:
			return
		}
	}

	clientIDStr := fmt.Sprintf("%s", s.state.ClientID)
	req := storage.ListDataBlocksRequest_builder{
		ClientId: proto.String(clientIDStr),
	}

	md := metadata.New(map[string]string{
		"authorization": s.state.Token,
	})

	ctx = metadata.NewOutgoingContext(ctx, md)
	stream, err := s.client.ListDataBlocks(ctx, req.Build())
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case errChan <- err:
			return
		default:
			return
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				blockResp, err := stream.Recv()
				if err != nil {
					select {
					case <-ctx.Done():
						return
					case errChan <- err:
						return
					default:
						return
					}
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

				select {
				case <-ctx.Done():
					return
				case blocksChan <- blocks:
					s.offlineRepository.ReplaceSyncedBlocks(blocks)
					// return
				default:
					return
				}
			}
		}
	}()
}

func (s *clientStorageService) Ping(ctx context.Context) error {
	md := metadata.Pairs("authorization", s.state.Token)
	ctx = metadata.NewOutgoingContext(ctx, md)
	req := storage.PingRequest_builder{}.Build()
	_, err := s.client.Ping(ctx, req)

	return err
}

func (s *clientStorageService) StartupFromFile() error {
	if err := s.offlineRepository.RestoreFromFile(); err != nil {
		return err
	}

	state := s.offlineRepository.GetAppState()
	claims, err := utils.ParseUnverifiedJWT(state.Token)
	if err != nil {
		return err
	}

	state.UserID = claims.UserID
	s.state = state

	return nil
}

func (s *clientStorageService) UnloadToFile() error {
	return s.offlineRepository.SaveToFile()
}

func (s *clientStorageService) SyncOfflineBlocks(ctx context.Context) error {
	blocksCount := s.offlineRepository.CountUnsyncedBlocks()
	if blocksCount == 0 {
		return nil
	}

	chans := make([]chan error, 0, blocksCount)
	blocks := s.offlineRepository.ReadUnsyncedBlocks()
	for _, b := range blocks {
		ch := make(chan error)
		chans = append(chans, ch)
		go s.SaveBlock(ctx, *b, ch)
	}

	for i, ch := range chans {
		err := <-ch
		if err != nil {
			return err
		}
		s.offlineRepository.DeleteBlockByData(blocks[i].Data)
	}

	return nil
}
