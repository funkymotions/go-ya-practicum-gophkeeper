package grpc

import (
	"context"
	"errors"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type authGRPCServer struct {
	auth.UnimplementedAuthServiceServer
	authService ports.AuthService
}

func NewAuthGRPCServer(authService ports.AuthService) *authGRPCServer {
	return &authGRPCServer{
		authService: authService,
	}
}

func (s *authGRPCServer) Register(
	ctx context.Context,
	req *auth.RegisterRequest,
) (*auth.RegisterResponse, error) {
	username, password := req.GetUsername(), req.GetPassword()
	token, err := s.authService.Register(username, password)
	var apperr *apperror.AppError
	if errors.As(err, &apperr) {
		appErr := err.(*apperror.AppError)
		return nil, status.Errorf(appErr.GRPCStatus, "%s", appErr.Message)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	resp := auth.RegisterResponse_builder{
		Token: proto.String(token),
	}.Build()

	return resp, nil
}

func (s *authGRPCServer) Authenticate(
	ctx context.Context,
	req *auth.AuthRequest,
) (*auth.AuthResponse, error) {
	username, password := req.GetUsername(), req.GetPassword()
	token, err := s.authService.Authenticate(username, password)
	var apperr *apperror.AppError
	if errors.As(err, &apperr) {
		appErr := err.(*apperror.AppError)
		return nil, status.Errorf(appErr.GRPCStatus, "%s", appErr.Message)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	resp := auth.AuthResponse_builder{
		Token: proto.String(token),
	}.Build()

	return resp, nil
}
