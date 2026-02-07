package client

import (
	"context"
	"fmt"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/subscription"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type clientSubscriptionService struct {
	client subscription.SubscriptionServiceClient
	state  *types.State
}

var _ ports.ClientSubscriber = (*clientSubscriptionService)(nil)

func NewClientSubscriptionService(state *types.State, client subscription.SubscriptionServiceClient) *clientSubscriptionService {
	return &clientSubscriptionService{
		state:  state,
		client: client,
	}
}

func (s *clientSubscriptionService) Subscribe() error {
	clientIDStr := fmt.Sprintf("%s", s.state.ClientID)
	md := metadata.New(map[string]string{
		"authorization": s.state.Token,
	})

	ctx := context.Background()
	ctx = metadata.NewOutgoingContext(ctx, md)
	sub := subscription.SubscribeRequest_builder{
		ClientId: proto.String(clientIDStr),
	}.Build()

	_, err := s.client.Subscribe(ctx, sub)
	if err != nil {
		return err
	}

	return nil
}

func (s *clientSubscriptionService) Unsubscribe() error {
	clientIDStr := fmt.Sprintf("%s", s.state.ClientID)
	md := metadata.New(map[string]string{
		"authorization": s.state.Token,
	})

	ctx := context.Background()
	ctx = metadata.NewOutgoingContext(ctx, md)
	unsub := subscription.UnsubscribeRequest_builder{
		ClientId: proto.String(clientIDStr),
	}.Build()

	_, err := s.client.Unsubscribe(ctx, unsub)
	if err != nil {
		return err
	}

	return nil
}
