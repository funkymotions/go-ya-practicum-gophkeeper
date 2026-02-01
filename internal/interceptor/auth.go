package interceptor

import (
	"context"
	"fmt"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type UserIDKey string

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func validateToken(m metadata.MD, secret []byte) (int, error) {
	token := m.Get("authorization")
	var tokenString string
	if len(token) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "missing authorization token")
	} else {
		tokenString = token[0]
	}

	claims, err := utils.CheckJWTToken(tokenString, secret)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return claims.UserID, nil
}

func StreamAuthInterceptor(secret []byte) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Errorf(codes.Internal, "missing metadata")
		}

		userID, err := validateToken(md, secret)
		if err != nil {
			return err
		}

		newCtx := context.WithValue(ss.Context(), UserIDKey("userID"), userID)
		wrappedSrvStreamCtx := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrappedSrvStreamCtx)
	}
}

func UnaryAuthInterceptor(secret []byte) grpc.UnaryServerInterceptor {
	var authEntrypointsToSkip = map[string]struct{}{
		"/auth.AuthService/Register":     {},
		"/auth.AuthService/Authenticate": {},
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		fmt.Printf("intercepting method: %s\n", info.FullMethod)
		if _, ok := authEntrypointsToSkip[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "missing metadata")
		}

		userID, err := validateToken(md, secret)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, UserIDKey("userID"), userID)

		return handler(ctx, req)
	}
}
