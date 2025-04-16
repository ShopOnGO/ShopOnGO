package home

import (
	"html/template"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/configs"
	_ "github.com/ShopOnGO/ShopOnGO/docs"
	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/gorilla/mux"
)

type HomeHandlerDeps struct { //  DC
	Config *configs.Config
	*HomeService
}
type HomeHandler struct {
	Config *configs.Config
	*HomeService
}

func NewHomeHandler(router *mux.Router, deps HomeHandlerDeps) {
	handler := &HomeHandler{
		Config:      deps.Config,
		HomeService: deps.HomeService,
	}
	//router.HandleFunc("GET /{hash}", handler.GoTo())
	//router.Handle("GET /link", middleware.IsAuthed(handler.GetAll(), deps.Config))
	router.Handle("/home", middleware.AuthOrGuest(handler.GetHomePage(), deps.Config)).Methods("GET")
}

// func (h *HomeHandler) GetAll() http.Handler {
// 	panic("unimplemented")
// }

// func (h *HomeHandler) GoTo() func(w http.ResponseWriter, r *http.Request) {
// 	panic("unimplemented")
// }

// GetHomePage возвращает данные для главной страницы магазина
// @Summary        Главная страница
// @Description    Получает информацию о популярных товарах, категориях потом и акциях
// @Tags          home
// @Produce       json
// @Success       200 {object} HomeData
// @Router        /home [get]
// GetHomePage возвращает главную страницу магазина с использованием шаблона
// @Summary        Главная страница
// @Description    Получает информацию о популярных товарах, категориях и акциях и отображает их через HTML-шаблон
// @Tags          home
// @Router        /home [get]

func (h *HomeHandler) GetHomePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Данные, которые будут переданы в шаблон
		userID := r.Context().Value(middleware.ContextUserIDKey)

		// Путь к шаблону
		tmplPath := "static/templates/home.html"

		// Загружаем и рендерим шаблон
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Рендерим шаблон и передаем данные
		err = tmpl.Execute(w, userID)
		if err != nil {
			http.Error(w, "Ошибка рендеринга шаблона: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
