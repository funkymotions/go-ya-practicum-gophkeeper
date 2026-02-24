package app

import (
	"fmt"
	"log"
	"net"

	apigrpc "github.com/funkymotions/go-ya-practicum-gophkeeper/internal/api/grpc"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/config"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/infrastructure/database"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/interceptor"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
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
	isDoneChan  chan struct{}
}

func New(config *config.Config, isDone chan struct{}) *Application {
	app := &Application{}
	app.conf = &config.Server
	dbConfig := config.Database
	app.isDoneChan = isDone
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

	blocksSqlRepo := repository.NewSQLRepository[model.Block](db)
	typesSqlRepo := repository.NewSQLRepository[model.Type](db)
	usersSqlRepo := repository.NewSQLRepository[model.User](db)

	// repositories
	storageRepository := repository.NewBlockRepository(blocksSqlRepo)
	subscriptionRepository := repository.NewSubscriptionRepository()
	typesRepository := repository.NewTypeRepository(typesSqlRepo)
	userRepository := repository.NewUserRepository(usersSqlRepo)

	// services
	subscriptionService := service.NewSubscriptionService(subscriptionRepository, app.logger)
	storageService := service.NewStorageService(
		service.StorageServiceArgs{
			StorageRepository:   storageRepository,
			SubscriptionService: subscriptionService,
			TypesRepository:     typesRepository,
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

	typesService := service.NewTypesService(typesRepository, app.logger)

	// gRPC servers
	grpcStorageServer := apigrpc.NewStorageGRPCServer(
		storageService,
		subscriptionService,
		typesService,
	)

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
	a.logger.Infow("Starting gRPC server", "host", a.conf.Host, "port", a.conf.Port)
	addr := fmt.Sprintf("%s:%d", a.conf.Host, a.conf.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return a.srv.Serve(l)
}

func (app *Application) Shutdown() {
	app.srv.Stop()
	app.isDoneChan <- struct{}{}
	app.logger.Info("gRPC servers stopped")
}
