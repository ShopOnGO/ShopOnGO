package migrations

import (
	"os"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/internal/cart"
	"github.com/ShopOnGO/ShopOnGO/internal/category"
	"github.com/ShopOnGO/ShopOnGO/internal/chat"
	"github.com/ShopOnGO/ShopOnGO/internal/favorites"
	"github.com/ShopOnGO/ShopOnGO/internal/link"
	"github.com/ShopOnGO/ShopOnGO/internal/product"
	"github.com/ShopOnGO/ShopOnGO/internal/productVariant"
	"github.com/ShopOnGO/ShopOnGO/internal/question"
	"github.com/ShopOnGO/ShopOnGO/internal/review"
	"github.com/ShopOnGO/ShopOnGO/internal/stat"
	"github.com/ShopOnGO/ShopOnGO/internal/user"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func CheckForMigrations() error {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		logger.Info("🚀 Starting migrations...")
		if err := RunMigrations(); err != nil {
			logger.Errorf("Error processing migrations: %v", err)
		}
	}
	return nil
}

func RunMigrations() error {
	// Загружаем конфиг (берёт env напрямую, а локально ещё и из .env если есть)
	cfg := configs.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.Db.Dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(
		&link.Link{},
		&stat.Stat{},
		&user.User{},
		&product.Product{}, &productVariant.ProductVariant{},
		&category.Category{},
		&brand.Brand{},
		&cart.Cart{}, &cart.CartItem{}, &favorites.Favorite{},
		&review.Review{}, &question.Question{},
		&chat.Message{},
	)

	if err != nil {
		return err
	}

	logger.Info("✅ Migrations completed successfully")
	return nil
}
