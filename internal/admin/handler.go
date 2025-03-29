package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	pb "github.com/ShopOnGO/admin-proto/pkg/service"
	"google.golang.org/grpc"
)

type AdminHandler struct {
	Clients *GRPCClients
}

func InitGRPCClients() *GRPCClients {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Ошибка подключения к gRPC серверу: %v", err)
	}

	return &GRPCClients{
		CategoryClient: pb.NewCategoryServiceClient(conn), //done
		BrandClient:    pb.NewBrandServiceClient(conn),    //done
		LinkClient:     pb.NewLinkServiceClient(conn),     //done
		ProductClient:  pb.NewProductServiceClient(conn),  //done
		UserClient:     pb.NewUserServiceClient(conn),     //done
		StatClient:     pb.NewStatServiceClient(conn),     //done
		HomeClient:     pb.NewHomeServiceClient(conn),     //done
	}
}

func NewAdminHandler(router *http.ServeMux) {
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
	// router.HandleFunc("DELETE /admin/users/all", handler.DeleteAllUsers)

	// Products
	router.HandleFunc("POST /admin/products", handler.CreateProduct)
	router.HandleFunc("GET /admin/products/featured", handler.GetFeaturedProducts)
	router.HandleFunc("PUT /admin/products", handler.UpdateProduct)
	router.HandleFunc("DELETE /admin/products", handler.DeleteProduct)
	//router.HandleFunc("DELETE /admin/products/all", handler.DeleteAllProducts)

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
// @Param          category body pb.CreateCategoryRequest true "Данные для создания категории"
// @Success        201 {object} pb.Category "Созданная категория"
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
// @Summary        Получение категории
// @Description    Возвращает категорию по переданному ID
// @Tags           category
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Param          id query int true "ID категории"
// @Success        200 {object} pb.Category "Найденная категория"
// @Failure        400 {string} string "Некорректный ID"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/categories [get]
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
// @Success        200 {array} pb.Category "Список категорий"
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
// @Param         brand body CreateBrandRequest true "Данные для создания бренда"
// @Success       201 {object} Brand
// @Failure       400 {string} string "Некорректный запрос"
// @Router        /admin/brands[post]
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
// @Summary        Рекомендованные бренды
// @Description    Возвращает список популярных или продвигаемых брендов
// @Tags           brand
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Success        200 {array} pb.Brand "Список брендов"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/brands/featured [get]
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
// @Summary        Рекомендованные продукты
// @Description    Возвращает список популярных или продвигаемых продуктов
// @Tags           product
// @Accept         json
// @Produce        json
// @Security       ApiKeyAuth
// @Success        200 {array} pb.Product "Список продуктов"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /admin/products/featured [get]
func (a *AdminHandler) GetFeaturedProducts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Clients.ProductClient.GetFeaturedProducts(ctx, &pb.FeaturedRequest{
		Amount:         5,
		Random:         true,
		IncludeDeleted: false,
	})
	if err != nil {
		http.Error(w, "Failed to get products", http.StatusInternalServerError)
		return
	}
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
