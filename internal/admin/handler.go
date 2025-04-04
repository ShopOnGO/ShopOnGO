package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	pb "github.com/ShopOnGO/admin-proto/pkg/service"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

type AdminHandler struct {
	Clients *GRPCClients
}

func InitGRPCClients() *GRPCClients {
	conn, err := grpc.Dial("admin_container:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Ошибка подключения к gRPC серверу: %v", err)
	}
	fmt.Println("grpc connected")
	return &GRPCClients{
		CategoryClient:       pb.NewCategoryServiceClient(conn), //done
		BrandClient:          pb.NewBrandServiceClient(conn),    //done
		LinkClient:           pb.NewLinkServiceClient(conn),     //done
		ProductClient:        pb.NewProductServiceClient(conn),  //done
		UserClient:           pb.NewUserServiceClient(conn),     //done
		StatClient:           pb.NewStatServiceClient(conn),     //done
		HomeClient:           pb.NewHomeServiceClient(conn),     //done
		ProductVariantClient: pb.NewProductVariantServiceClient(conn),
	}
}

func NewAdminHandler(router *mux.Router) {
	handler := &AdminHandler{
		Clients: InitGRPCClients(),
	}
	//Home
	router.HandleFunc("GET /home", handler.GetHomeData)

	//Stat
	router.HandleFunc("POST /stats/click", handler.AddClick)

	// Users
	router.HandleFunc("POST /admin/users", handler.CreateUser)
	router.HandleFunc("GET /admin/users", handler.GetUserByEmail)
	router.HandleFunc("PUT /admin/users", handler.UpdateUser)
	router.HandleFunc("DELETE /admin/users", handler.DeleteUser)
	router.HandleFunc("POST /admin/by-email", handler.GetUserByEmail)
	// router.HandleFunc("DELETE /admin/users/all", handler.DeleteAllUsers)

	//ProductVariants
	router.HandleFunc("/admin/products/{product_id}/variants/add", handler.CreateProductVariant).Methods("POST")
	router.HandleFunc("/admin/products/{product_id}/find", handler.GetVariant).Methods("POST")
	router.HandleFunc("/admin/products/{product_id}/variants", handler.ListVariants).Methods("POST")
	router.HandleFunc("/admin/products/{product_id}/variants/{id}", handler.UpdateProductVariant).Methods("PUT")
	router.HandleFunc("/admin/products/{product_id}/variants/{id}/stock", handler.ManageStock).Methods("POST")
	router.HandleFunc("/admin/products/{product_id}/variants/{id}", handler.DeleteVariant).Methods("DELETE")

	// Products
	router.HandleFunc("/admin/products", handler.CreateProduct).Methods("POST")
	router.HandleFunc("/admin/products/featured", handler.GetFeaturedProducts).Methods("GET")
	router.HandleFunc("/admin/products", handler.UpdateProduct).Methods("PUT")
	router.HandleFunc("/admin/products", handler.DeleteProduct).Methods("DELETE")
	// router.HandleFunc("/admin/products/all", handler.DeleteAllProducts).Methods("DELETE")
	//Дописать как реализовать?
	// при нажатии нужно получать сразу все варианты, и потом пользователь уже будет с кешем этих данных разбираться, что ему нужно.
	//

	//Brands
	router.HandleFunc("POST /admin/brands", handler.CreateBrand)
	router.HandleFunc("GET /admin/brands/featured", handler.GetFeaturedBrands)
	router.HandleFunc("PUT /admin/brands", handler.UpdateBrand)
	router.HandleFunc("DELETE /admin/brands", handler.DeleteBrand)
	//router.HandleFunc("DELETE /admin/brands/all", handler.DeleteAllBrands)

	// Categories
	router.HandleFunc("POST /admin/categories", handler.CreateCategory)
	router.HandleFunc("GET /admin/categories/featured", handler.GetFeaturedCategories)
	router.HandleFunc("PUT /admin/categories/{id}", handler.UpdateCategory) // {id}???????????
	router.HandleFunc("DELETE /admin/categories/", handler.DeleteCategory)
	//router.HandleFunc("DELETE /admin/categories/all", handler.DeleteAllCategories)

}

// CreateCategory создаёт новую категорию
// @Summary        Создание категории
// @Description    Добавляет новую категорию в базу данных
// @Tags           category
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// Param category body pb.CreateCategoryRequest true "Данные для создания категории"
// @Success 201 {object} pb.Category "Созданная категория"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/categories [post]
func (a *AdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.CategoryClient.CreateCategory(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.Category)
}

// GetCategoryByID получает категорию по её ID
// @Summary      Получение категории
// @Description  Возвращает категорию по переданному ID
// @Tags         category
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id  path  int  true  "ID категории"
// @Success      200 {object} pb.Category "Найденная категория"
// @Failure      400 {string} string "Некорректный ID"
// @Failure      500 {string} string "Ошибка сервера"
// @Router       /admin/categories/{id} [get]
func (a *AdminHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	categoryID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.CategoryClient.FindCategoryByID(ctx, &pb.FindCategoryByIDRequest{Id: uint32(categoryID)})
	if err != nil {
		http.Error(w, "Failed to fetch category", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.Category)
}

// UpdateCategory обновляет существующую категорию
// @Summary        Обновление категории
// @Description    Изменяет имя и/или описание категории
// @Tags           category
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id path int true "ID категории"
// @Param          category body pb.UpdateCategoryRequest true "Данные для обновления категории"
// @Success        200 {object} pb.Category "Обновлённая категория"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        404 {string} string "Категория не найдена"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/categories/{id} [put]
func (a *AdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	categoryID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var req pb.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Id = uint32(categoryID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.CategoryClient.UpdateCategory(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.Category)
}

// GetFeaturedCategories получает список рекомендованных категорий
// @Summary        Рекомендованные категории
// @Description    Возвращает список популярных или продвигаемых категорий
// @Tags           category
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Success        200 {array}  pb.Category "Список категорий"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/categories/featured [get]
func (a *AdminHandler) GetFeaturedCategories(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.CategoryClient.GetFeaturedCategories(ctx, &pb.GetFeaturedCategoriesRequest{Amount: 5})
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.Categories)
}

// DeleteCategory удаляет категорию по имени
// @Summary        Удаление категории
// @Description    Удаляет существующую категорию из базы данных
// @Tags           category
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          category body pb.DeleteCategoryByNameRequest true "Данные для удаления категории"
// @Success        200 {string} string "Категория успешно удалена"
// @Failure        400 {string} string "Некорректное имя"
// @Failure        404 {string} string "Категория не найдена"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/categories [delete]
func (a *AdminHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	var req pb.DeleteCategoryByNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := a.Clients.CategoryClient.DeleteCategory(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteAllCategories удаляет все категории
// @Summary        Удаление всех категорий
// @Description    Удаляет все существующие категории из базы данных без возможности восстановления
// @Tags           category
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Success        200 {string} string "Все категории успешно удалены"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/categories/all [delete]
func (a *AdminHandler) DeleteAllCategories(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.CategoryClient.GetFeaturedCategories(ctx, &pb.GetFeaturedCategoriesRequest{Amount: 0, Unscoped: true})
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	for _, category := range resp.Categories {
		_, err := a.Clients.CategoryClient.DeleteCategory(ctx, &pb.DeleteCategoryByNameRequest{Name: category.Name, Unscoped: true})
		if err != nil {
			log.Printf("❌ Error deleting category Name=%s: %v", category.Name, err)
		}
	}
	w.WriteHeader(http.StatusOK)
}

// Create создает новый бренд
// @Summary        Новый бренд
// @Description    Создает новый бренд по имени и заносит его в базу
// @Tags          brand
// @Accept        json
// @Produce       json
// @Security      ApiKeyAuth
// @Param         brand body pb.CreateBrandRequest true "Данные для создания бренда"
// @Success       201 {object} pb.Brand
// @Failure       400 {string} string "Некорректный запрос"
// @Router        /admin/brands [post]
func (a *AdminHandler) CreateBrand(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateBrandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.BrandClient.CreateBrand(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to create brand", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.Brand)
}

// GetFeaturedBrands получает список рекомендованных брендов
// @Summary      Рекомендованные бренды
// @Description  Возвращает список популярных или продвигаемых брендов
// @Tags         brand
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        amount    query  int  false  "Количество брендов"
// @Param        unscoped  query  bool false  "Показывать архивные бренды"
// @Success      200 {array} pb.Brand "Список брендов"
// @Failure      500 {string} string "Ошибка сервера"
// @Router       /admin/brands/featured [get]
func (a *AdminHandler) GetFeaturedBrands(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.BrandClient.GetFeaturedBrands(ctx, &pb.GetFeaturedBrandsRequest{Amount: 5, Unscoped: true})
	if err != nil {
		http.Error(w, "Failed to get brands", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.Brands)
}

// UpdateBrand обновляет информацию о бренде
// @Summary        Обновление бренда
// @Description    Изменяет данные существующего бренда
// @Tags           brand
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          brand body pb.Brand true "Данные для обновления бренда"
// @Success        200 {object} pb.Brand "Обновлённый бренд"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/brands [put]
func (a *AdminHandler) UpdateBrand(w http.ResponseWriter, r *http.Request) {
	var req pb.Brand
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.BrandClient.UpdateBrand(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to update brand", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.Brand)
}

// DeleteBrand удаляет бренд
// @Summary        Удаление бренда
// @Description    Удаляет существующий бренд из базы данных
// @Tags           brand
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          brand body pb.DeleteBrandRequest true "Данные для удаления бренда"
// @Success        200 {string} string "Бренд успешно удалён"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        404 {string} string "Бренд не найден"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/brands [delete]
func (a *AdminHandler) DeleteBrand(w http.ResponseWriter, r *http.Request) {
	var req pb.DeleteBrandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := a.Clients.BrandClient.DeleteBrand(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to delete brand", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteAllBrands удаляет все бренды
// @Summary        Удаление всех брендов
// @Description    Удаляет все бренды из базы данных
// @Tags           brand
// @Security       ApiKeyAuth
// @Success        200 {string} string "Все бренды успешно удалены"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/brands/all [delete]
func (a *AdminHandler) DeleteAllBrands(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.BrandClient.GetFeaturedBrands(ctx, &pb.GetFeaturedBrandsRequest{Amount: 0, Unscoped: true})
	if err != nil {
		http.Error(w, "Failed to retrieve brands", http.StatusInternalServerError)
		return
	}

	for _, brand := range resp.Brands {
		_, err := a.Clients.BrandClient.DeleteBrand(ctx, &pb.DeleteBrandRequest{Name: brand.Name, Unscoped: true})
		if err != nil {
			continue // Можно логировать ошибку
		}
	}
	w.WriteHeader(http.StatusOK)
}

// CreateProduct создает новый продукт
// @Summary        Новый продукт
// @Description    Создает новый продукт по имени и заносит его в базу
// @Tags           product
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          product body pb.Product true "Данные для создания продукта"
// @Success        201 {object} pb.Product "Созданный продукт"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/products [post]
func (a *AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req pb.Product
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.ProductClient.CreateProduct(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.Product)
}

// GetFeaturedProducts получает список рекомендованных продуктов
// @Summary      Рекомендованные продукты
// @Description  Возвращает список популярных или продвигаемых продуктов
// @Tags         product
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        amount          query  int  false  "Количество продуктов"
// @Param        random          query  bool false  "Случайный порядок"
// @Param        includeDeleted  query  bool false  "Включать удалённые продукты"
// @Success      200 {array} pb.ProductList "Список продуктов"
// @Failure      500 {string} string "Ошибка сервера"
// @Router       /admin/products/featured [get]
func (a *AdminHandler) GetFeaturedProducts(w http.ResponseWriter, r *http.Request) {
	// Получаем Query-параметры
	query := r.URL.Query()

	amount, err := strconv.Atoi(query.Get("amount"))
	if err != nil || amount <= 0 {
		amount = 5 // Значение по умолчанию
	}

	random, _ := strconv.ParseBool(query.Get("random"))
	includeDeleted, _ := strconv.ParseBool(query.Get("include_deleted"))

	// Создаем gRPC-запрос с параметрами
	req := &pb.FeaturedRequest{
		Amount:         uint32(amount),
		Random:         random,
		IncludeDeleted: includeDeleted,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.ProductClient.GetFeaturedProducts(ctx, req)
	if err != nil {
		http.Error(w, "Failed to get products", http.StatusInternalServerError)
		return
	}

	// Отправляем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Products)
}

// UpdateProduct обновляет существующий продукт
// @Summary        Обновление продукта
// @Description    Изменяет данные продукта
// @Tags           product
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          product body pb.Product true "Данные для обновления продукта"
// @Success        200 {object} pb.Product "Обновленный продукт"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/products [put]
func (a *AdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var req pb.Product
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.ProductClient.UpdateProduct(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.Product)
}

// DeleteProduct удаляет продукт
// @Summary        Удаление продукта
// @Description    Удаляет существующий продукт из базы данных
// @Tags           product
// @Security       ApiKeyAuth
// @Param          product body pb.DeleteProductRequest true "Данные для удаления продукта"
// @Success        200 {string} string "Продукт успешно удален"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/products [delete]
func (a *AdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	var req pb.DeleteProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := a.Clients.ProductClient.DeleteProduct(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteAllProducts удаляет все продукты
// @Summary        Удаление всех продуктов
// @Description    Удаляет все продукты из базы данных
// @Tags           product
// @Security       ApiKeyAuth
// @Success        200 {string} string "Все продукты успешно удалены"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/products/all [delete]
func (a *AdminHandler) DeleteAllProducts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.ProductClient.GetFeaturedProducts(ctx, &pb.FeaturedRequest{Amount: 0, IncludeDeleted: true})
	if err != nil {
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	for _, product := range resp.Products {
		_, err := a.Clients.ProductClient.DeleteProduct(ctx, &pb.DeleteProductRequest{Id: uint64(product.Model.Id), Unscoped: true})
		if err != nil {
			continue // Можно логировать ошибку
		}
	}
	w.WriteHeader(http.StatusOK)
}

// CreateUser создает нового пользователя
// @Summary        Новый пользователь
// @Description    Создает нового пользователя по имени и заносит его в базу
// @Tags           user
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          user body pb.User true "Данные для создания пользователя"
// @Success        201 {object} pb.User "Созданный пользователь"
// @Failure        400 {string} string "Некорректный запрос"
// @Router         /admin/users [post]
func (a *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req pb.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.UserClient.CreateUser(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.User)
}

// GetUserByEmail возвращает пользователя по email
// @Summary        Получение пользователя по email
// @Description    Поиск пользователя в базе данных по его email
// @Tags           user
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          request body pb.EmailRequest true "Email пользователя для поиска"
// @Success        200 {object} pb.User "Найденный пользователь"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        404 {string} string "Пользователь не найден"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/users/by-email [post]
func (a *AdminHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	var req pb.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.UserClient.FindUserByEmail(ctx, &req)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(resp.User)
}

// UpdateUser обновляет существующий продукт
// @Summary        Обновление продукта
// @Description    Изменяет данные продукта
// @Tags           users
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          users body pb.User true "Данные для обновления пользователя"
// @Success        200 {object} pb.User "Обновленный пользователь"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/users [put]
func (a *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req pb.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.UserClient.UpdateUser(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp.User)
}

// DeleteUser удаляет пользователя
// @Summary        Удаление пользователя
// @Description    Удаляет существующего пользователя из базы данных
// @Tags           users
// @Security       ApiKeyAuth
// @Param          users body pb.DeleteUserRequest true "Данные для удаления пользователя"
// @Success        200 {string} string "пользователь успешно удален"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/users [delete]
func (a *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var req pb.DeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := a.Clients.UserClient.DeleteUser(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteAllUsers удаляет всех пользователей
// @Summary        Удаление всех пользователей
// @Description    Удаляет всех пользователей из базы данных
// @Tags           users
// @Security       ApiKeyAuth
// @Success        200 {string} string "Все пользователи успешно удалены"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/users/all [delete]
func (a *AdminHandler) DeleteAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users := []string{"test@example.com", "test-updated@example.com"}

	for _, email := range users {
		resp, err := a.Clients.UserClient.FindUserByEmail(ctx, &pb.EmailRequest{Email: email})
		if err != nil {
			continue // Можно логировать ошибку
		}

		_, err = a.Clients.UserClient.DeleteUser(ctx, &pb.DeleteUserRequest{Id: uint64(resp.User.Model.Id), Unscoped: true})
		if err != nil {
			continue // Можно логировать ошибку
		}
	}
	w.WriteHeader(http.StatusOK)
}

// AddClick добавляет клик по элементу
// @Summary        Добавление клика
// @Description    Добавляет информацию о клике по элементу
// @Tags           stats
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          click body pb.ClickRequest true "Данные клика"
// @Success        200 {string} string "Клик успешно добавлен"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/stats/click [post]
func (a *AdminHandler) AddClick(w http.ResponseWriter, r *http.Request) {
	var req pb.ClickRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := a.Clients.StatClient.AddClick(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to add click", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Click added successfully"))
}

// GetHomeData получает данные для главной страницы
// @Summary        Получение данных для главной
// @Description    Получает информацию, необходимую для отображения главной страницы
// @Tags           home
// @Produce        json
// @Security       ApiKeyAuth
// @Success        200 {object} pb.HomeDataResponse "Данные для главной страницы"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/home [get]
func (a *AdminHandler) GetHomeData(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.HomeClient.GetHomeData(ctx, &pb.EmptyRequest{})
	if err != nil {
		http.Error(w, "Failed to retrieve home data", http.StatusInternalServerError)
		return
	}

	if resp.Error != nil {
		http.Error(w, resp.Error.Message, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateProductVariant создает новый вариант продукта
// @Summary        Создание варианта продукта
// @Description    Создает новый вариант для указанного продукта
// @Tags           productVariant
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          product_id path int true "ID продукта"
// @Param          variant body pb.ProductVariant true "Данные варианта"
// @Success        201 {object} pb.ProductVariant "Созданный вариант"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        404 {string} string "Продукт не найден"
// @Failure        409 {string} string "Конфликт данных"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/products/{product_id}/variants/add [post]
func (a *AdminHandler) CreateProductVariant(w http.ResponseWriter, r *http.Request) {

	// Получаем параметры из URL
	vars := mux.Vars(r)
	productIDStr, ok := vars["product_id"]
	if !ok {
		http.Error(w, `{"error": "Missing product_id"}`, http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil || productID == 0 {
		http.Error(w, `{"error": "Invalid product ID"}`, http.StatusBadRequest)
		return
	}

	// Декодируем тело запроса
	var req pb.ProductVariant
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Проверяем обязательные поля
	if req.Sku == "" {
		http.Error(w, `{"error": "SKU is required"}`, http.StatusBadRequest)
		return
	}

	// Устанавливаем product_id из URL в структуру запроса
	req.ProductId = uint32(productID)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	// Отправляем запрос в gRPC-сервис
	resp, err := a.Clients.ProductVariantClient.CreateVariant(ctx, &req)
	if err != nil {
		st, _ := status.FromError(err)
		var statusCode int
		var message string

		switch st.Code() {
		case codes.NotFound:
			statusCode = http.StatusNotFound
			message = "Product not found"
		case codes.AlreadyExists:
			statusCode = http.StatusConflict
			message = "Variant with this SKU already exists"
		case codes.InvalidArgument:
			statusCode = http.StatusBadRequest
			message = st.Message()
		default:
			statusCode = http.StatusInternalServerError
			message = "Failed to create variant"
		}

		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, message), statusCode)
		return
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp.Variant); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// GetVariant возвращает вариант продукта по SKU, Barcode или ID
// @Summary        Получение варианта продукта
// @Description    Поиск варианта продукта по одному из идентификаторов (SKU, Barcode или ID)
// @Tags           productVariant
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          request body pb.VariantRequest true "Идентификатор для поиска (SKU, Barcode или ID)"
// @Success        200 {object} pb.ProductVariant "Найденный вариант продукта"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        404 {string} string "Вариант не найден"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/products/{product_id}/find [post]
func (a *AdminHandler) GetVariant(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из URL
	vars := mux.Vars(r)
	productID, err := strconv.ParseUint(vars["product_id"], 10, 32)
	if err != nil || productID == 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req pb.VariantRequest

	// Используем protojson для декодирования тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := protojson.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Identifier == nil {
		http.Error(w, "At least one identifier must be provided", http.StatusBadRequest)
		return
	}

	// Можно использовать productID, если он нужен
	// req.ProductId = uint32(productID) // если есть поле ProductId в proto

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Отправляем запрос в gRPC-сервис
	resp, err := a.Clients.ProductVariantClient.GetVariant(ctx, &req)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			http.Error(w, "Product variant not found", http.StatusNotFound)
		case codes.InvalidArgument:
			http.Error(w, st.Message(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if resp == nil || resp.Variant == nil {
		http.Error(w, "Empty service response", http.StatusInternalServerError)
		return
	}

	// Ответ в обычном JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp.Variant); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// UpdateProductVariant обновляет существующий вариант продукта
// @Summary        Обновление варианта продукта
// @Description    Изменяет данные варианта продукта по идентификатору
// @Tags           productVariant
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          product_id  path int            true "ID продукта"
// @Param          id          path int            true "ID варианта продукта"
// @Param          variant     body pb.ProductVariant true "Данные для обновления варианта"
// @Success        200 {object} pb.ProductVariant "Обновленный вариант продукта"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        404 {string} string "Вариант не найден"
// @Failure        409 {string} string "Конфликт данных"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/{product_id}/variants/{id} [put]
func (a *AdminHandler) UpdateProductVariant(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из URL
	vars := mux.Vars(r)
	productID, err := strconv.ParseUint(vars["product_id"], 10, 32)
	if err != nil || productID == 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	variantID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil || variantID == 0 {
		http.Error(w, "Invalid variant ID", http.StatusBadRequest)
		return
	}

	var req pb.ProductVariant

	// Декодируем JSON в структуру
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем, что Model не nil
	if req.Model == nil {
		req.Model = &pb.Model{} // Создаем экземпляр Model, если его нет
	}

	// Заполняем недостающие поля
	req.Model.Id = uint32(variantID)
	req.ProductId = uint32(productID)

	// Теперь req полностью заполнен и можно его использовать

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Выполняем обновление
	resp, err := a.Clients.ProductVariantClient.UpdateVariant(ctx, &req)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			http.Error(w, "Product variant not found", http.StatusNotFound)
		case codes.InvalidArgument:
			http.Error(w, st.Message(), http.StatusBadRequest)
		case codes.AlreadyExists:
			http.Error(w, "SKU or barcode conflict", http.StatusConflict)
		default:
			logger.Info(err)
			http.Error(w, "Failed to update variant", http.StatusInternalServerError)
		}
		return
	}

	if resp == nil || resp.Variant == nil {
		http.Error(w, "Empty service response", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp.Variant); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// ManageStock обрабатывает операции с запасами варианта продукта
// @Summary        Управление запасами
// @Description    Выполняет операции резервирования, освобождения и обновления стока
// @Tags           productVariant
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          product_id path int            true "ID продукта"
// @Param          id         path int            true "ID варианта продукта"
// @Param          request    body pb.StockRequest true "Параметры операции с запасами"
// @Success        204        "Операция выполнена успешно"
// @Failure        400        {object} pb.Error "Некорректные параметры запроса"
// @Failure        404        {object} pb.Error "Вариант не найден"
// @Failure        409        {object} pb.Error "Конфликт при выполнении операции"
// @Failure        500        {object} pb.Error "Внутренняя ошибка сервера"
// @Router         /admin/products/{product_id}/variants/{id}/stock [post]
func (a *AdminHandler) ManageStock(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из URL
	vars := mux.Vars(r)

	productID, err := strconv.ParseUint(vars["product_id"], 10, 32)
	if err != nil || productID == 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	variantID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil || variantID == 0 {
		http.Error(w, "Invalid variant ID", http.StatusBadRequest)
		return
	}

	var req pb.StockRequest

	// Декодируем JSON в структуру
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Устанавливаем ID варианта
	req.VariantId = uint32(variantID)

	// Контекст с таймаутом
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Отправляем запрос в gRPC-сервис
	_, err = a.Clients.ProductVariantClient.ManageStock(ctx, &req)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			http.Error(w, "Product variant not found", http.StatusNotFound)
		case codes.InvalidArgument:
			http.Error(w, st.Message(), http.StatusBadRequest)
		case codes.FailedPrecondition:
			http.Error(w, st.Message(), http.StatusConflict)
		default:
			http.Error(w, "Failed to manage stock", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListVariants возвращает список вариантов продуктов с фильтрами
// @Summary        Получение списка вариантов с фильтрами
// @Description    Возвращает список вариантов продуктов с возможностью фильтрации
// @Tags           productVariant
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          product_id   path int     true  "ID продукта"
// @Param          active_only  query bool   false "Только активные варианты"
// @Param          price_min    query number false "Минимальная цена"
// @Param          price_max    query number false "Максимальная цена"
// @Param          limit        query int    false "Лимит записей"
// @Param          offset       query int    false "Смещение"
// @Success        200 {object} pb.VariantListResponse "Список вариантов"
// @Failure        400 {string} string "Некорректные параметры запроса"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/products/{product_id}/variants [get]
func (a *AdminHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	// Извлекаем product_id из URL
	pathRe := regexp.MustCompile(`^/admin/products/(\d+)/variants$`)
	matches := pathRe.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil || productID <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	req := &pb.VariantListRequest{
		ProductId:  uint32(productID),
		PriceRange: &pb.PriceRange{},
	}

	// Парсинг остальных параметров
	req.ActiveOnly = query.Has("active_only")

	// Обработка ценового диапазона
	if minPrice := query.Get("price_min"); minPrice != "" {
		if min, err := strconv.ParseFloat(minPrice, 32); err == nil {
			req.PriceRange.Min = uint32(min)
		} else {
			http.Error(w, "Invalid price_min format", http.StatusBadRequest)
			return
		}
	}

	if maxPrice := query.Get("price_max"); maxPrice != "" {
		if max, err := strconv.ParseFloat(maxPrice, 32); err == nil {
			req.PriceRange.Max = uint32(max)
		} else {
			http.Error(w, "Invalid price_max format", http.StatusBadRequest)
			return
		}
	}

	// Парсинг пагинации
	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.ParseInt(limit, 10, 32); err == nil {
			req.Limit = uint32(l)
		} else {
			http.Error(w, "Invalid limit format", http.StatusBadRequest)
			return
		}
	}

	if offset := query.Get("offset"); offset != "" {
		if o, err := strconv.ParseInt(offset, 10, 32); err == nil {
			req.Offset = uint32(o)
		} else {
			http.Error(w, "Invalid offset format", http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.ProductVariantClient.ListVariants(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.InvalidArgument:
			http.Error(w, st.Message(), http.StatusBadRequest)
		case codes.NotFound:
			http.Error(w, "Product not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to list variants", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// DeleteVariant выполняет мягкое или полное удаление варианта продукта
// @Summary        Удаление варианта продукта
// @Description    Выполняет мягкое (по умолчанию) или полное удаление варианта
// @Tags           productVariant
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          product_id path int  true  "ID продукта"
// @Param          id         path int  true  "ID варианта"
// @Param          unscoped   query bool false "Полное удаление из базы (без возможности восстановления)"
// @Success        204 "Удаление выполнено успешно"
// @Failure        400 {string} string "Некорректные параметры"
// @Failure        404 {string} string "Вариант не найден"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/products/{product_id}/variants/{id} [delete]
func (a *AdminHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из URL
	vars := mux.Vars(r)

	// Парсим product_id
	productID, err := strconv.ParseInt(vars["product_id"], 10, 64)
	if err != nil || productID <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Парсим variant_id
	variantID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil || variantID <= 0 {
		http.Error(w, "Invalid variant ID", http.StatusBadRequest)
		return
	}

	// Структура для парсинга тела запроса
	var body struct {
		Unscoped bool `json:"unscoped"`
	}

	// Читаем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем gRPC-запрос
	req := &pb.DeleteVariantRequest{
		Id:       uint32(variantID),
		Unscoped: body.Unscoped,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Выполняем удаление
	_, err = a.Clients.ProductVariantClient.DeleteVariant(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			http.Error(w, "Variant not found", http.StatusNotFound)
		case codes.InvalidArgument:
			http.Error(w, st.Message(), http.StatusBadRequest)
		default:
			http.Error(w, "Failed to delete variant", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
