package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/interceptor"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/storage"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type storageGRPCServer struct {
	storage.UnimplementedStorageServiceServer
	storageService      ports.StorageService
	subscriptionService ports.SubscriptionService
	typesService        ports.TypesService
	isAppExiting        chan struct{}
	isDone              chan struct{}
}

func NewStorageGRPCServer(
	service ports.StorageService,
	subscriptionService ports.SubscriptionService,
	typesService ports.TypesService,
) *storageGRPCServer {
	return &storageGRPCServer{
		storageService:      service,
		subscriptionService: subscriptionService,
		typesService:        typesService,
	}
}

func (s *storageGRPCServer) SaveDataBlock(
	ctx context.Context,
	req *storage.SaveDataBlockRequest,
) (*storage.SaveDataBlockResponse, error) {
	userID := ctx.Value(interceptor.UserIDKey)
	userIDInt, ok := userID.(int)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	block := &model.Block{
		UserID:  userIDInt,
		Title:   req.GetTitle(),
		Data:    req.GetChiphertext(),
		Salt:    req.GetSalt(),
		Nonce:   req.GetNonce(),
		Profile: req.GetProfile().String(),
		TypeID:  int(req.GetTypeId()),
	}

	block, err := s.storageService.SaveDataBlock(userIDInt, block)
	if err != nil {
		return nil, err
	}

	return storage.SaveDataBlockResponse_builder{}.Build(), nil
}

func (s *storageGRPCServer) ListDataBlocks(
	req *storage.ListDataBlocksRequest,
	stream storage.StorageService_ListDataBlocksServer,
) error {
	ctx := stream.Context()

	// hold connection for 10 minutes to reduce stale connections
	// then, a new subscription can be to fetch a data updates
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	userID := ctx.Value(interceptor.UserIDKey)
	clientID := req.GetClientId()
	userIDInt, ok := userID.(int)
	if !ok {
		return status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	send := func() error {
		blocks, err := s.storageService.ListDataBlocks(userIDInt)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to list data blocks: %v", err)
		}

		var respBlocks []*storage.DataBlock
		for _, block := range blocks {
			blockBuilder := storage.DataBlock_builder{
				BlockId:     proto.Int32(int32(block.ID)),
				Title:       proto.String(block.Title),
				Chiphertext: block.Data,
				Salt:        block.Salt,
				Nonce:       block.Nonce,
				Profile:     storage.EncProfile(utils.ProfileToProto(utils.ScryptProfile(block.Profile))).Enum(),
				Type: storage.BlockType_builder{
					Id:          proto.Int32(int32(block.Type.ID)),
					TypeName:    proto.String(block.Type.TypeName),
					Description: proto.String(block.Type.Description),
				}.Build(),
			}

			respBlock := blockBuilder.Build()
			respBlocks = append(respBlocks, respBlock)
		}

		b := storage.ListDataBlocksResponse_builder{
			DataBlocks: respBlocks,
		}

		resp := b.Build()

		return stream.Send(resp)
	}

	subs := s.subscriptionService.GetUserSubscribers(userIDInt)
	if len(subs) == 0 {
		return status.Errorf(codes.NotFound, "no subscribers found for user %d", userIDInt)
	}
	if err := send(); err != nil {
		return err
	}

	for {
		select {
		case <-ctxWithTimeout.Done():
			// unsubscribe the client
			s.subscriptionService.Unsubscribe(userIDInt, clientID)
			return ctxWithTimeout.Err()
		case <-subs[clientID]:
			if err := send(); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *storageGRPCServer) ListBlockTypes(
	ctx context.Context,
	req *storage.GetBlockTypesRequest,
) (*storage.GetBlockTypesResponse, error) {
	types, err := s.typesService.ReadAllTypes()
	var appError *apperror.AppError
	if err != nil && errors.As(err, &appError) {
		return nil, status.Errorf(appError.GRPCStatus, "%s", appError.Message)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list block types: %v", err)
	}

	var respTypes []*storage.BlockType
	for _, t := range types {
		typeBuilder := storage.BlockType_builder{
			Id:          proto.Int32(int32(t.ID)),
			TypeName:    proto.String(t.TypeName),
			Description: proto.String(t.Description),
		}

		respType := typeBuilder.Build()
		respTypes = append(respTypes, respType)
	}

	b := storage.GetBlockTypesResponse_builder{
		BlockTypes: respTypes,
	}

	resp := b.Build()

	return resp, nil
}

func (s *storageGRPCServer) Ping(
	ctx context.Context,
	req *storage.PingRequest,
) (*storage.PingResponse, error) {
	return &storage.PingResponse{}, nil
}
