package grpc

import (
	"sync"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/auth"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/storage"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/subscription"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	AuthClient         auth.AuthServiceClient
	StorageClient      storage.StorageServiceClient
	SubscriptionClient subscription.SubscriptionServiceClient
}

var once sync.Once
var client *GRPCClient

func NewGRPCClient(conn ...*grpc.ClientConn) *GRPCClient {
	once.Do(func() {
		client = &GRPCClient{
			AuthClient:         auth.NewAuthServiceClient(conn[0]),
			StorageClient:      storage.NewStorageServiceClient(conn[0]),
			SubscriptionClient: subscription.NewSubscriptionServiceClient(conn[0]),
		}
	})

	return client
}
