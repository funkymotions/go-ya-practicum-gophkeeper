package app

import (
	"fmt"
	"log"
	"net"

	apigrpc "github.com/funkymotions/go-ya-practicum-gophkeeper/internal/api/grpc"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/config"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/infrastructure/database"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/interceptor"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/auth"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/storage"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/subscription"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/repository"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type services struct {
	storageServer      storage.StorageServiceServer
	authServer         auth.AuthServiceServer
	subscriptionServer subscription.SubscriptionServiceServer
}

type Application struct {
	grpcServers services
	conf        *config.Server
	srv         *grpc.Server
	logger      *zap.SugaredLogger
}

// TODO: add logger
// TODO: add graceful shutdown
// TODO: add migrations
func New(config *config.Config) *Application {
	app := &Application{}
	app.conf = &config.Server
	dbConfig := config.Database
	db, err := database.NewSQLDriver(&dbConfig)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	defer l.Sync()
	app.logger = l.Sugar()

	// repositories
	storageRepository := repository.NewStorageRepository(db)
	userRepository := repository.NewUserRepository(db)
	subscriptionRepository := repository.NewSubscriptionRepository()

	// services
	subscriptionService := service.NewSubscriptionService(subscriptionRepository)
	storageService := service.NewStorageService(
		service.StorageServiceArgs{
			StorageRepository:   storageRepository,
			SubscriptionService: subscriptionService,
			Logger:              app.logger,
		},
	)

	authService := service.NewAuthService(
		service.AuthServiceArgs{
			UserRepository: userRepository,
			Logger:         app.logger,
			JWTSecret:      []byte(config.Server.JWT.SecretKey),
		},
	)

	// gRPC servers
	grpcStorageServer := apigrpc.NewStorageGRPCServer(storageService, subscriptionService)
	grpcAuthServer := apigrpc.NewAuthGRPCServer(authService)
	grpcSubscriptionServer := apigrpc.NewSubscriptionGRPCServer(subscriptionService)
	app.grpcServers.storageServer = grpcStorageServer
	app.grpcServers.authServer = grpcAuthServer
	app.grpcServers.subscriptionServer = grpcSubscriptionServer

	app.srv = grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryAuthInterceptor([]byte(config.Server.JWT.SecretKey))),
		grpc.StreamInterceptor(interceptor.StreamAuthInterceptor([]byte(config.Server.JWT.SecretKey))),
	)

	auth.RegisterAuthServiceServer(app.srv, app.grpcServers.authServer)
	storage.RegisterStorageServiceServer(app.srv, app.grpcServers.storageServer)
	subscription.RegisterSubscriptionServiceServer(app.srv, app.grpcServers.subscriptionServer)

	return app
}

func (a *Application) Start() error {
	log.Printf("Starting gRPC server...")
	addr := fmt.Sprintf("%s:%d", a.conf.Host, a.conf.Port)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return a.srv.Serve(l)
}

// TODO: make graceful shutdown
func (a *Application) Shudown() {
	a.srv.GracefulStop()
	log.Println("gRPC servers stopped")
}
