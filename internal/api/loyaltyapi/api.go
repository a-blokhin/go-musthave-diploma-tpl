package loyaltyapi

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/usecase/getbalanceusecase"
	"go-musthave-diploma-tpl/internal/usecase/getordersusecase"
	"go-musthave-diploma-tpl/internal/usecase/getwithdrawalsusecase"
	"go-musthave-diploma-tpl/internal/usecase/orderuploadusecase"
	"go-musthave-diploma-tpl/internal/usecase/userloginusecase"
	"go-musthave-diploma-tpl/internal/usecase/userregistrationusecase"
	"go-musthave-diploma-tpl/internal/usecase/withdrawusecase"
	"go-musthave-diploma-tpl/pkg/getbalancepkg"
	"go-musthave-diploma-tpl/pkg/getorderspkg"
	"go-musthave-diploma-tpl/pkg/getwithdrawalspkg"
	"go-musthave-diploma-tpl/pkg/orderuploadpkg"
	"go-musthave-diploma-tpl/pkg/userloginpkg"
	"go-musthave-diploma-tpl/pkg/userregisterpkg"
	"go-musthave-diploma-tpl/pkg/withdrawpkg"
)

type LoyaltyAPI struct {
	userRegistrationUseCase *userregistrationusecase.Usecase
	userLoginUseCase        *userloginusecase.Usecase
	orderUploadUseCase      *orderuploadusecase.Usecase
	getOrdersUseCase        *getordersusecase.Usecase
	getBalanceUseCase       *getbalanceusecase.Usecase
	withdrawUseCase         *withdrawusecase.Usecase
	getWithdrawalsUseCase   *getwithdrawalsusecase.Usecase
	jwtService              *auth.JWTService
	logger                  *zap.Logger
}

func New(
	userRegistrationUseCase *userregistrationusecase.Usecase,
	userLoginUseCase *userloginusecase.Usecase,
	orderUploadUseCase *orderuploadusecase.Usecase,
	getOrdersUseCase *getordersusecase.Usecase,
	getBalanceUseCase *getbalanceusecase.Usecase,
	withdrawUseCase *withdrawusecase.Usecase,
	getWithdrawalsUseCase *getwithdrawalsusecase.Usecase,
	jwtService *auth.JWTService,
	logger *zap.Logger,
) *LoyaltyAPI {
	return &LoyaltyAPI{
		userRegistrationUseCase: userRegistrationUseCase,
		userLoginUseCase:        userLoginUseCase,
		orderUploadUseCase:      orderUploadUseCase,
		getOrdersUseCase:        getOrdersUseCase,
		getBalanceUseCase:       getBalanceUseCase,
		withdrawUseCase:         withdrawUseCase,
		getWithdrawalsUseCase:   getWithdrawalsUseCase,
		jwtService:              jwtService,
		logger:                  logger,
	}
}

func (api *LoyaltyAPI) RegisterHandlers(router *gin.Engine) {
	// User authentication endpoints (no auth required)
	router.POST(userregisterpkg.MethodPath, api.userRegistrationUseCase.Execute)
	router.POST(userloginpkg.MethodPath, api.userLoginUseCase.Execute)

	// Protected routes group with JWT middleware
	protected := router.Group("/")
	protected.Use(middleware.JWTAuthMiddleware(api.jwtService, api.logger))
	{
		// Order endpoints (auth required)
		protected.POST(orderuploadpkg.MethodPath, api.orderUploadUseCase.Execute)
		protected.GET(getorderspkg.MethodPath, api.getOrdersUseCase.Execute)

		// Balance endpoints (auth required)
		protected.GET(getbalancepkg.MethodPath, api.getBalanceUseCase.Execute)
		protected.POST(withdrawpkg.MethodPath, api.withdrawUseCase.Execute)

		// Withdrawal history endpoint (auth required)
		protected.GET(getwithdrawalspkg.MethodPath, api.getWithdrawalsUseCase.Execute)
	}
}