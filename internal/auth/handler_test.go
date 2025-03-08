package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/auth"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/refresh"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/db"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func bootstrap() (*auth.AuthHandler, sqlmock.Sqlmock, error) {
	database, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	gormDb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database,
	}))
	userRepo := user.NewUserRepository(&db.Db{
		DB: gormDb,
	})
	authRepo := refresh.NewAuthRepository(&db.Db{
		DB: gormDb,
	})
	handler := auth.AuthHandler{
		Config: &configs.Config{
			Auth: configs.AuthConfig{
				Secret: "Secret",
			},
		}, AuthService: auth.NewAuthService(userRepo, authRepo),
	}
	return &handler, mock, nil
}

func TestLoginHandlerSuccess(t *testing.T) {
	handler, mock, err := bootstrap()
	rows := sqlmock.NewRows([]string{"email", "password"}).AddRow("sadtest123@d.ru", "$2a$10$Ijdf1/9t3f/OrAqq6U9/LeTD.NocZJigAPvyLrLbWBSsOf.lX3Mdi")
	mock.ExpectQuery("SELECT").WillReturnRows(rows) // select можно описать полным запросом, но так как мы вызываем только один раз нам пофигу на какой селект оно сработает
	// это мы имитировали работу поиска по емэил, нам похер на параметры и все такое,
	if err != nil {
		t.Fatal(err)
		return
	}
	data, _ := json.Marshal(&auth.LoginRequest{
		Email:    "sadtest123@d.ru",
		Password: "123",
	})
	//query
	reader := bytes.NewReader(data)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", reader)
	handler.Login()(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("got %d,Expected %d", w.Code, 200)
	}
}

// очень немало работы для интеграциооных тестов

func TestRegisterHandlerSuccess(t *testing.T) {
	handler, mock, err := bootstrap()
	rows := sqlmock.NewRows([]string{"email", "password", "name"})
	mock.ExpectQuery("SELECT").WillReturnRows(rows) // select можно описать полным запросом, но так как мы вызываем только один раз нам пофигу на какой селект оно сработает
	// это мы имитировали работу поиска по емэил, нам похер на параметры и все такое,
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1")) //ПОСМОТРЕВ НА ИСХОДНИК МЫ ДОЛЖНЫ ПОНЯТЬ ЧТО ВОЗВРАТ НАС НЕ ИНТЕРЕСУЕТ, ГЛАВНОЕ ЧТОБЫ БЕЗ ОШИБКИ!!!
	mock.ExpectCommit()
	if err != nil {
		t.Fatal(err)
		return
	}
	data, _ := json.Marshal(&auth.RegisterRequest{
		LoginRequest: auth.LoginRequest{ // Явно указываем, что поля Email и Password относятся к LoginRequest
			Email:    "sadtest123@d.ru",
			Password: "123",
		},
		Name: "dan",
	})
	//query
	reader := bytes.NewReader(data)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/register", reader)
	handler.Register()(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("got %d,Expected %d", w.Code, 201)
	}
}
