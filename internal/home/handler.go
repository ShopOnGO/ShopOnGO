package home

import (
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	_ "github.com/ShopOnGO/ShopOnGO/prod/docs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/res"
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
func (h *HomeHandler) GetHomePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { //ctx

		homeData, err := h.HomeService.GetHomeData()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Json(w, homeData, 200)
	}
}

//Generate
