package main

import (
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	_ "github.com/ShopOnGO/ShopOnGO/prod/docs"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/auth"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/category"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/home"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/link"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/product"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/stat"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"
	"github.com/ShopOnGO/ShopOnGO/prod/migrations"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/db"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/event"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/middleware"
)

func App() http.Handler {

	//AutoMigrate
	migrations.CheckForMigrations()

	conf := configs.LoadConfig()
	db := db.NewDB(conf)
	//cache := cache.NewRedis(conf)
	router := http.NewServeMux()
	eventBus := event.NewEventBus() // передаем как зависимость в handle

	// REPOSITORIES
	linkRepository := link.NewLinkRepository(db)
	userRepository := user.NewUserRepository(db)
	statRepository := stat.NewStatRepository(db)
	categoryRepository := category.NewCategoryRepository(db)
	productsRepository := product.NewProductRepository(db)

	// Services
	authService := auth.NewAuthService(userRepository)
	homeService := home.NewHomeService(categoryRepository, productsRepository)
	statService := stat.NewStatService(&stat.StatServiceDeps{
		StatRepository: statRepository,
		EventBus:       eventBus,
	})

	//Handlers
	auth.NewAuthHandler(router, auth.AuthHandlerDeps{
		Config:      conf,
		AuthService: authService,
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

	// swagger
	router.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./docs"))))

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
