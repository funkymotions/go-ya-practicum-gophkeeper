package client

import (
	"context"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/types"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/client/view"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/config"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/auth"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/storage"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/subscription"
	clientrepo "github.com/funkymotions/go-ya-practicum-gophkeeper/internal/repository/client"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/service/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type services struct {
	authService    ports.ClientAuthService
	storageService ports.ClientStorageService
	sub            ports.ClientSubscriber
}

type ClientApp struct {
	rootModel        tea.Model
	program          *tea.Program
	grpcConn         *grpc.ClientConn
	appChan          chan os.Signal
	isOnline         chan bool
	blocksChan       chan []*model.Block
	blocksErrChan    chan error
	blocks           *[]*model.Block
	state            *types.State
	services         services
	listenerChan     chan struct{}
	tickerChan       chan struct{}
	streamChan       chan struct{}
	isListenerDone   chan struct{}
	isTickerDone     chan struct{}
	isStreamDone     chan struct{}
	blocksUpdateChan chan struct{}
	errUpdateChan    chan error
	isStreamStarted  bool
	buildDate        string
	clientVersion    string
	ctx              context.Context
	cancelFn         context.CancelFunc
}

func NewClientApp(
	c config.ClientConf,
	g *grpc.ClientConn,
	appChan chan os.Signal,
	listenerChan chan struct{},
	tickerChan chan struct{},
	streamChan chan struct{},
	isListenerDone chan struct{},
	isTickerDone chan struct{},
	isStreamDone chan struct{},
	buildDate string,
	clientVersion string,
) *ClientApp {
	return &ClientApp{
		grpcConn:         g,
		appChan:          appChan,
		state:            types.NewState(),
		isOnline:         make(chan bool, 2),
		blocksChan:       make(chan []*model.Block, 1),
		blocksErrChan:    make(chan error, 1),
		blocks:           &[]*model.Block{},
		blocksUpdateChan: make(chan struct{}, 1),
		errUpdateChan:    make(chan error, 1),
		listenerChan:     listenerChan,
		tickerChan:       tickerChan,
		streamChan:       streamChan,
		isListenerDone:   isListenerDone,
		isTickerDone:     isTickerDone,
		isStreamDone:     isStreamDone,
		ctx:              nil,
		cancelFn:         nil,
		buildDate:        buildDate,
		clientVersion:    clientVersion,
	}
}

func (app *ClientApp) Init() error {
	storageGRPCClient := storage.NewStorageServiceClient(app.grpcConn)
	subGRPCClient := subscription.NewSubscriptionServiceClient(app.grpcConn)
	authGRPCClient := auth.NewAuthServiceClient(app.grpcConn)

	// offline
	offlineRepo := clientrepo.NewOfflineRepository(app.state)

	// services
	app.services.authService = client.NewClientAuthService(authGRPCClient)
	app.services.sub = client.NewClientSubscriptionService(app.state, subGRPCClient)
	app.services.storageService = client.NewClientStorageService(
		client.ClientStorageServiceArgs{
			Client:            storageGRPCClient,
			State:             app.state,
			OfflineRepository: offlineRepo,
		},
	)

	// models
	regModel := view.NewAuthModel(
		view.AuthModelArgs{
			Service:  app.services.authService,
			ViewType: view.RegisterViewType,
			State:    app.state,
		},
	)

	authModel := view.NewAuthModel(
		view.AuthModelArgs{
			Service:  app.services.authService,
			ViewType: view.AuthViewType,
			State:    app.state,
		},
	)

	storageModel := view.NewStorageModel(
		view.StorageModelArgs{
			State: app.state,
			AddBlockModel: view.NewAddBlockView(
				view.AddBlockViewArgs{
					State:   app.state,
					Service: app.services.storageService,
				},
			),
			ListBlockModel: view.NewBlockListView(
				view.BlockListViewArgs{
					State:       app.state,
					Service:     app.services.storageService,
					Blocks:      app.blocks,
					UpdateCh:    app.blocksUpdateChan,
					ErrUpdateCh: app.errUpdateChan,
				},
			),
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	app.ctx = ctx
	app.cancelFn = cancel
	app.RestoreAppState()
	app.pollClientState()
	go app.listenForOnline()
	app.StartBlockFetching()
	app.rootModel = view.NewRootModel(
		view.RootModelArgs{
			AppChan:        app.appChan,
			RegisterModel:  regModel,
			AuthModel:      authModel,
			StorageModel:   storageModel,
			State:          app.state,
			StorageService: app.services.storageService,
			BuildDate:      app.buildDate,
			ClientVersion:  app.clientVersion,
		},
	)

	return nil
}

func (app *ClientApp) RestoreAppState() error {
	err := app.services.storageService.StartupFromFile()
	if err != nil {
		return err
	}

	return nil
}

func (app *ClientApp) listenForOnline() {
	ctx, cancel := context.WithCancel(context.Background())
	for {
		select {
		case <-app.listenerChan:
			cancel()
			app.isListenerDone <- struct{}{}
			return
		case isOnline := <-app.isOnline:
			if isOnline {
				err := app.services.storageService.SyncOfflineBlocks(ctx)
				if err != nil {
					continue
				}
			}
		}
	}
}

func (app *ClientApp) StartBlockFetching() {
	go func() {
		for {
			select {
			case <-app.streamChan:
				app.cancelFn()
				// streaming stopped, exit goroutine
				if app.state.IsOnline {
					app.services.sub.Unsubscribe()
				}
				app.isStreamDone <- struct{}{}
				return
			case blocks := <-app.blocksChan:
				*app.blocks = blocks
				select {
				case app.blocksUpdateChan <- struct{}{}:
				default:
				}
			case err := <-app.blocksErrChan:
				select {
				case app.errUpdateChan <- err:
				default:
				}
			case isOnline := <-app.isOnline:
				if app.isStreamStarted {
					continue
				}
				if !isOnline {
					// retrieve offline blocks from a service
					go app.services.storageService.StartBlockStream(
						app.ctx,
						app.blocksChan,
						app.blocksErrChan,
					)
					continue
				}

				err := app.services.sub.Subscribe()
				if err != nil {
					continue
				}

				app.isStreamStarted = true
				go app.services.storageService.StartBlockStream(
					app.ctx,
					app.blocksChan,
					app.blocksErrChan,
				)
			}
		}
	}()
}

func (app *ClientApp) pollClientState() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		app.pingServer()
		for {
			select {
			case <-app.tickerChan:
				app.isTickerDone <- struct{}{}
				return
			case <-ticker.C:
				err := app.pingServer()
				if err != nil {
					continue
				}
			}
		}
	}()
}

func (app *ClientApp) Start() error {
	app.program = tea.NewProgram(app.rootModel)
	if _, err := app.program.Run(); err != nil {
		return err
	}

	return nil
}

func (app *ClientApp) Close() error {
	app.program.Quit()

	return app.grpcConn.Close()
}

func (app *ClientApp) pingServer() error {
	err := app.services.storageService.Ping(context.Background())
	if err != nil {
		app.state.IsOnline = false
		app.isOnline <- false
		app.isOnline <- false

		return err
	}

	isOnline := app.grpcConn.GetState() == connectivity.Ready
	app.state.IsOnline = isOnline

	// send isOffline twice
	app.isOnline <- isOnline
	app.isOnline <- isOnline

	return nil
}
