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
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/internal/admin"
	"github.com/ShopOnGO/ShopOnGO/internal/auth"
	"github.com/ShopOnGO/ShopOnGO/internal/auth/passwordreset"
	"github.com/ShopOnGO/ShopOnGO/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/internal/cart"
	"github.com/ShopOnGO/ShopOnGO/internal/category"
	"github.com/ShopOnGO/ShopOnGO/internal/chat"
	"github.com/ShopOnGO/ShopOnGO/internal/home"
	"github.com/ShopOnGO/ShopOnGO/internal/link"
	"github.com/ShopOnGO/ShopOnGO/internal/notification"
	"github.com/ShopOnGO/ShopOnGO/internal/product"
	"github.com/ShopOnGO/ShopOnGO/internal/productVariant"
	"github.com/ShopOnGO/ShopOnGO/internal/question"
	"github.com/ShopOnGO/ShopOnGO/internal/review"
	"github.com/ShopOnGO/ShopOnGO/internal/stat"
	"github.com/ShopOnGO/ShopOnGO/internal/user"

	"github.com/ShopOnGO/ShopOnGO/migrations"
	"github.com/ShopOnGO/ShopOnGO/pkg/db"
	"github.com/ShopOnGO/ShopOnGO/pkg/event"
	"github.com/ShopOnGO/ShopOnGO/pkg/kafkaService"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/pkg/oauth2"
	"github.com/ShopOnGO/ShopOnGO/pkg/redisdb"
	"github.com/gorilla/mux"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/ShopOnGO/ShopOnGO/docs"
)

func App() http.Handler {

	//AutoMigrate
	migrations.CheckForMigrations()

	conf := configs.LoadConfig()
	db := db.NewDB(conf)
	redis := redisdb.NewRedisDB(conf)
	router := mux.NewRouter()
	eventBus := event.NewEventBus() // передаем как зависимость в handle
	kafkaProducers := kafkaService.InitKafkaProducers(
		conf.Kafka.Brokers,
		conf.Kafka.Topics,
	)

	// REPOSITORIES
	linkRepository := link.NewLinkRepository(db)
	userRepository := user.NewUserRepository(db)
	statRepository := stat.NewStatRepository(db)
	chatRepository := chat.NewChatRepository(db)
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
	chatService := chat.NewChatService(chatRepository)
	statService := stat.NewStatService(&stat.StatServiceDeps{
		StatRepository: statRepository,
		EventBus:       eventBus,
	})

	oauth2Service := oauth2.NewOAuth2Service(conf, refreshTokenRepository)
	resetService := passwordreset.NewResetService(conf, resetPasswordRepository, userRepository, kafkaProducers["reset"])

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
		Config:  conf,
	})
	passwordreset.NewResetHandler(router, passwordreset.ResetHandlerDeps{
		ResetService: resetService,
		Config:       conf,
	})

	review.NewReviewHandler(router, review.ReviewHandlerDeps{
		Kafka:  kafkaProducers["reviews"],
		Config: conf,
	})
	question.NewQuestionHandler(router, question.QuestionHandlerDeps{
		Kafka:  kafkaProducers["reviews"],
		Config: conf,
	})
	notification.NewNotificationHandler(router, notification.NotificationHandlerDeps{
		Kafka:  kafkaProducers["notifications"],
		Config: conf,
	})
	product.NewProductHandler(router, product.ProductHandlerDeps{
		Kafka:  kafkaProducers["products"],
		Config: conf,
	})
	productVariant.NewProductVariantHandler(router, productVariant.ProductVariantHandlerDeps{
		Kafka:  kafkaProducers["productVariants"],
		Config: conf,
	})

	chat.NewChatHandler(router, chat.ChatHandlerDeps{
		ChatService: chatService,
		Config:      conf,
	})
	admin.NewAdminHandler(router)

	// swagger
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	//обработчик подписки ( бесконечно сидит отдельно и ждёт пока не придут сообщения)
	go statService.AddClick()

	//Middlewares
	stack := middleware.Chain(
		middleware.CORS,
		middleware.Logging,
	)

	// Обработка статических файлов (например, /static/js/notifications.js)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

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
