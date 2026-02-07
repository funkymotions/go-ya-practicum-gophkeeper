package grpc

import (
	"context"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/interceptor"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/subscription"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type subscriptionGRPCServer struct {
	subscription.UnimplementedSubscriptionServiceServer
	subscriptionService ports.SubscriptionService
}

func NewSubscriptionGRPCServer(subscriptionService ports.SubscriptionService) *subscriptionGRPCServer {
	return &subscriptionGRPCServer{
		subscriptionService: subscriptionService,
	}
}

func (s *subscriptionGRPCServer) Subscribe(
	ctx context.Context,
	req *subscription.SubscribeRequest,
) (*subscription.SubscribeResponse, error) {
	userIDRaw := ctx.Value(interceptor.UserIDKey)
	userID, ok := userIDRaw.(int)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	s.subscriptionService.Subscribe(userID, req.GetClientId())

	return &subscription.SubscribeResponse{}, nil
}

func (s *subscriptionGRPCServer) Unsubscribe(
	ctx context.Context,
	req *subscription.UnsubscribeRequest,
) (*subscription.UnsubscribeResponse, error) {
	userIDRaw := ctx.Value(interceptor.UserIDKey)
	userID, ok := userIDRaw.(int)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	s.subscriptionService.Unsubscribe(userID, req.GetClientId())

	return &subscription.UnsubscribeResponse{}, nil
}
