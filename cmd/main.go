// @title ShopOnGO API
// @version 1.0
// @description API сервиса ShopOnGO, обеспечивающего авторизацию, управление пользователями, товарами и аналитикой.
// @termsOfService http://shopongo.com/terms/

// @contact.name Support Team
// @contact.url http://shopongo.com/support
// @contact.email support@shopongo.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	_ "github.com/ShopOnGO/ShopOnGO/prod/docs"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/auth"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/auth/passwordreset"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/cart"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/category"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/home"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/link"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/product"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/stat"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"

	"github.com/ShopOnGO/ShopOnGO/prod/migrations"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/db"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/email/smtp"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/event"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/kafkaService"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/oauth2"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/redisdb"

	httpSwagger "github.com/swaggo/http-swagger"
)

func App() http.Handler {

	//AutoMigrate
	migrations.CheckForMigrations()

	conf := configs.LoadConfig()
	db := db.NewDB(conf)
	redis := redisdb.NewRedisDB(conf)
	router := http.NewServeMux()
	eventBus := event.NewEventBus() // передаем как зависимость в handle
	smtp := smtp.NewSMTPSender(conf.SMTP.Name, conf.SMTP.From, conf.SMTP.Pass, conf.SMTP.Host, conf.SMTP.Port)
	
	
	// Подключение к Kafka
	brokers := []string{"kafka:9092"}
	topic := "review-events"

	kafka := kafkaService.NewProducer(brokers, topic)
	defer kafka.Close()

	// Отправляем сообщение
	ctx := context.Background()
	err := kafka.Produce(ctx, []byte("review-id-123"), []byte(`{"action":"created","review_id":123}`))
	if err != nil {
		logger.Errorf("Ошибка при отправке сообщения в Kafka: %v", err)
	}
	logger.Info("✅ Сообщение отправлено в Kafka")


	// REPOSITORIES
	linkRepository := link.NewLinkRepository(db)
	userRepository := user.NewUserRepository(db)
	statRepository := stat.NewStatRepository(db)
	categoryRepository := category.NewCategoryRepository(db)
	productRepository := product.NewProductRepository(db)
	brandsRepository := brand.NewBrandRepository(db)
	cartRepository := cart.NewCartRepository(db)
	refreshTokenRepository := oauth2.NewRedisRefreshTokenRepository(redis)
	resetPasswordRepository := passwordreset.NewRedisResetRepository(redis)


	// Services
	authService := auth.NewAuthService(userRepository)
	homeService := home.NewHomeService(categoryRepository, productRepository, brandsRepository)
	cartService := cart.NewCartService(cartRepository)
	statService := stat.NewStatService(&stat.StatServiceDeps{
		StatRepository: statRepository,
		EventBus:       eventBus,
	})

	oauth2Service := oauth2.NewOAuth2Service(conf, refreshTokenRepository)
	resetService := passwordreset.NewResetService(conf, smtp, resetPasswordRepository, userRepository)
	//categoryService := category.NewCategoryService(categoryRepository)
	//brandService := brand.NewBrandService(brandsRepository)
	//statService := stat.NewStatService(statRepository)
	//prodService := product.NewProductService(productRepository)
	//userService := user.NewUserService(userRepository)

	//Handlers
	auth.NewAuthHandler(router, auth.AuthHandlerDeps{
		Config:        conf,
		AuthService:   authService,
		OAuth2Service: oauth2Service,
	})
	link.NewLinkHandler(router, link.LinkHandlerDeps{
		LinkRepository: linkRepository,
		EventBus:       eventBus,
		Config:         conf,
	})
	stat.NewStatHandler(router, stat.StatHandlerDeps{
		StatRepository: statRepository,
		Config:         conf,
	})
	home.NewHomeHandler(router, home.HomeHandlerDeps{
		HomeService: homeService,
		Config:      conf,
	})
	cart.NewCartHandler(router, cart.CartHandlerDeps{
		CartService: cartService,
		Config:      conf,
    })
	oauth2.NewOAuth2Handler(router, oauth2.OAuth2HandlerDeps{
		Service: oauth2Service,
		Config: conf,
	})
	passwordreset.NewResetHandler(router, passwordreset.ResetHandlerDeps{
		ResetService: resetService,
        Config:       conf,
	})

	// swagger
	router.Handle("/swagger/", httpSwagger.WrapHandler)

	//обработчик подписки ( бесконечно сидит отдельно и ждёт пока не придут сообщения)
	go statService.AddClick()

	//Middlewares
	stack := middleware.Chain(
		middleware.CORS,
		middleware.Logging,
	)
	return stack(router)
}

func main() {
	app := App()
	server := http.Server{
		Addr:    "0.0.0.0:8081",
		Handler: app,
	}

	logger.Info("Server started")
	server.ListenAndServe()

}
