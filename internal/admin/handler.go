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
	router.HandleFunc("GET /admin/users", handler.GetUserByEmail)
	router.HandleFunc("POST /admin/users", handler.CreateUser)
	router.HandleFunc("PUT /admin/users/{id}", handler.UpdateUser)
	router.HandleFunc("DELETE /admin/users/{id}", handler.DeleteUser)
	// router.HandleFunc("DELETE /admin/users", handler.DeleteAllUsers)

	// Products
	router.HandleFunc("POST /admin/products", handler.CreateProduct)
	router.HandleFunc("GET /admin/products", handler.GetFeaturedProducts)
	router.HandleFunc("PUT /admin/products/{id}", handler.UpdateProduct)
	router.HandleFunc("DELETE /admin/products/{id}", handler.DeleteProduct)
	//router.HandleFunc("DELETE /admin/products", handler.DeleteAllProducts)

	//Brands
	router.HandleFunc("POST /admin/brands", handler.CreateBrand)
	router.HandleFunc("GET /admin/brands", handler.GetFeaturedBrands)
	router.HandleFunc("PUT /admin/brands/{id}", handler.UpdateBrand)
	router.HandleFunc("DELETE /admin/brands/{id}", handler.DeleteBrand)
	//router.HandleFunc("DELETE /admin/brands", handler.DeleteAllBrands)

	// Categories
	router.HandleFunc("POST /admin/categories", handler.CreateCategory)
	router.HandleFunc("GET /admin/categories", handler.GetFeaturedCategories)
	router.HandleFunc("PUT /admin/categories/{id}", handler.UpdateCategory)
	router.HandleFunc("DELETE /admin/categories/{id}", handler.DeleteCategory)
	//router.HandleFunc("DELETE /admin/categories", handler.DeleteAllCategories)

}

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
