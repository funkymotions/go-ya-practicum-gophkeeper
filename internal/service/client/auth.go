package client

import (
	"context"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/auth"
	"google.golang.org/protobuf/proto"
)

type clientAuthService struct {
	client auth.AuthServiceClient
}

var _ ports.ClientAuthService = (*clientAuthService)(nil)

func NewClientAuthService(c auth.AuthServiceClient) *clientAuthService {
	return &clientAuthService{
		client: c,
	}
}

func (c *clientAuthService) Authenticate(username, password string) (string, error) {
	request := &auth.AuthRequest_builder{
		Username: proto.String(username),
		Password: proto.String(password),
	}

	resp, err := c.client.Authenticate(context.Background(), request.Build())
	if err != nil {
		return "", err
	}

	return resp.GetToken(), nil
}

func (c *clientAuthService) Register(username, password string) (string, error) {
	req := auth.RegisterRequest_builder{
		Username: proto.String(username),
		Password: proto.String(password),
	}

	resp, err := c.client.Register(context.Background(), req.Build())
	if err != nil {
		return "", err
	}

	return resp.GetToken(), nil
}
