package migrations

import (
	"os"

	"github.com/ShopOnGO/ShopOnGO/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/internal/cart"
	"github.com/ShopOnGO/ShopOnGO/internal/category"
	"github.com/ShopOnGO/ShopOnGO/internal/comment"
	"github.com/ShopOnGO/ShopOnGO/internal/link"
	"github.com/ShopOnGO/ShopOnGO/internal/product"
	"github.com/ShopOnGO/ShopOnGO/internal/productVariant"
	"github.com/ShopOnGO/ShopOnGO/internal/review"
	"github.com/ShopOnGO/ShopOnGO/internal/stat"
	"github.com/ShopOnGO/ShopOnGO/internal/user"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func CheckForMigrations() error {

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		logger.Info("🚀 Starting migrations...")
		if err := RunMigrations(); err != nil {
			logger.Errorf("Error processing migrations: %v", err)
		}
		return nil
	}
	// if not "migrate" args[1]
	return nil
}

func RunMigrations() error {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("DSN")), &gorm.Config{
		//DisableForeignKeyConstraintWhenMigrating: true, //временно игнорировать миграции в первый раз а потом их добавить
	})
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
		&cart.Cart{}, &cart.CartItem{}, 
		&review.Review{}, comment.Comment{},)

	if err != nil {
		return err
	}

	logger.Info("✅")
	return nil
}
