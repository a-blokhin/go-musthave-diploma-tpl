package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/accrual"
	"go-musthave-diploma-tpl/internal/api/loyaltyapi"
	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/database"
	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/migration"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/repository/postgresrepository"
	"go-musthave-diploma-tpl/internal/service"
	"go-musthave-diploma-tpl/internal/usecase/getbalanceusecase"
	"go-musthave-diploma-tpl/internal/usecase/getordersusecase"
	"go-musthave-diploma-tpl/internal/usecase/getwithdrawalsusecase"
	"go-musthave-diploma-tpl/internal/usecase/orderuploadusecase"
	"go-musthave-diploma-tpl/internal/usecase/userloginusecase"
	"go-musthave-diploma-tpl/internal/usecase/userregistrationusecase"
	"go-musthave-diploma-tpl/internal/usecase/withdrawusecase"
)

type DI struct {
	router *gin.Engine
	api    *loyaltyapi.LoyaltyAPI
	config *config.Config
	logger *zap.Logger
	db     *database.DB

	usecases struct {
		userRegistration *userregistrationusecase.Usecase
		userLogin        *userloginusecase.Usecase
		orderUpload      *orderuploadusecase.Usecase
		getOrders        *getordersusecase.Usecase
		getBalance       *getbalanceusecase.Usecase
		withdraw         *withdrawusecase.Usecase
		getWithdrawals   *getwithdrawalsusecase.Usecase
	}

	repos struct {
		userRepo       repository.UserRepository
		orderRepo      repository.OrderRepository
		balanceRepo    repository.BalanceRepository
		withdrawalRepo repository.WithdrawalRepository
	}

	services struct {
		jwtService      *auth.JWTService
		passwordService *auth.PasswordService
		accrualClient   *accrual.Client
		orderProcessor  *service.OrderProcessor
	}

	httpServer *http.Server
}

func (d *DI) Init(config *config.Config) error {
	d.config = config

	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := loggerConfig.Build()
	if err != nil {
		return err
	}
	d.logger = logger

	if config.DatabaseDSN != "" {
		db, err := database.New(config.DatabaseDSN)
		if err != nil {
			return err
		}
		d.db = db
		d.logger.Info("Connected to PostgreSQL database")
	}

	d.initRepos()
	d.initServices()
	d.initUsecases()
	d.initMux()
	d.initAPI()

	return nil
}

func (d *DI) initRepos() {
	if d.db != nil {
		d.logger.Info("Using PostgreSQL database storage")

		migrator := migration.New(d.logger, "migrations")
		if err := migrator.Up(d.config.DatabaseDSN); err != nil {
			d.logger.Fatal("Failed to run database migrations", zap.Error(err))
		}

		d.repos.userRepo = postgresrepository.NewUserRepository(d.db.Pool())
		d.repos.orderRepo = postgresrepository.NewOrderRepository(d.db.Pool())
		d.repos.balanceRepo = postgresrepository.NewBalanceRepository(d.db.Pool())
		d.repos.withdrawalRepo = postgresrepository.NewWithdrawalRepository(d.db.Pool())

	}
}

func (d *DI) initServices() {

	jwtSecret := d.config.JWTSecret
	if jwtSecret == "" {
		d.logger.Fatal("JWT_SECRET environment variable or -j flag is required")
	}
	d.services.jwtService = auth.NewJWTService(jwtSecret)

	d.services.passwordService = auth.DefaultPasswordService()

	if d.config.AccrualSystemAddress != "" {
		d.services.accrualClient = accrual.NewClient(d.config.AccrualSystemAddress, d.logger)
		d.logger.Info("Accrual client initialized", zap.String("address", d.config.AccrualSystemAddress))
	}
}

func (d *DI) initUsecases() {

	d.usecases.userRegistration = userregistrationusecase.New(
		d.repos.userRepo,
		d.services.passwordService,
		d.services.jwtService,
		d.repos.balanceRepo,
		d.logger,
	)

	d.usecases.userLogin = userloginusecase.New(
		d.repos.userRepo,
		d.services.passwordService,
		d.services.jwtService,
		d.logger,
	)

	d.usecases.orderUpload = orderuploadusecase.New(
		d.repos.orderRepo,
		d.logger,
	)

	if d.services.accrualClient != nil {
		d.services.orderProcessor = service.NewOrderProcessor(
			d.repos.orderRepo,
			d.repos.balanceRepo,
			d.services.accrualClient,
			d.logger,
		)
		d.logger.Info("Order processor initialized")
	}

	d.usecases.getOrders = getordersusecase.New(
		d.repos.orderRepo,
		d.logger,
	)

	d.usecases.getBalance = getbalanceusecase.New(
		d.repos.balanceRepo,
		d.logger,
	)

	d.usecases.withdraw = withdrawusecase.New(
		d.repos.balanceRepo,
		d.logger,
	)

	d.usecases.getWithdrawals = getwithdrawalsusecase.New(
		d.repos.withdrawalRepo,
		d.logger,
	)

}

func (d *DI) initMux() {
	gin.SetMode(gin.ReleaseMode)
	d.router = gin.New()
	d.router.Use(gin.Recovery())
	d.router.Use(middleware.GzipMiddleware())
	d.router.Use(middleware.LoggingMiddleware(d.logger))
}

func (d *DI) initAPI() {
	d.api = loyaltyapi.New(
		d.usecases.userRegistration,
		d.usecases.userLogin,
		d.usecases.orderUpload,
		d.usecases.getOrders,
		d.usecases.getBalance,
		d.usecases.withdraw,
		d.usecases.getWithdrawals,
		d.services.jwtService,
		d.logger,
	)
	d.api.RegisterHandlers(d.router)
}

func (d *DI) StartServer(ctx context.Context) error {
	d.httpServer = &http.Server{
		Addr:    d.config.ServerAddress,
		Handler: d.router,
	}

	if d.services.orderProcessor != nil {
		go d.services.orderProcessor.Start(ctx)
		d.logger.Info("Order processor started in background")
	}

	return d.httpServer.ListenAndServe()
}

func (d *DI) StopServer(ctx context.Context) error {

	if d.db != nil {
		d.db.Close()
	}

	if d.httpServer != nil {
		return d.httpServer.Shutdown(ctx)
	}
	return nil
}
