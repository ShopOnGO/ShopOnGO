package home

import (
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/res"
)

type HomeHandlerDeps struct { //  DC
	Config *configs.Config
	*HomeService
}
type HomeHandler struct {
	Config *configs.Config
	*HomeService
}

func NewHomeHandler(router *http.ServeMux, deps HomeHandlerDeps) {
	handler := &HomeHandler{
		Config:      deps.Config,
		HomeService: deps.HomeService,
	}
	router.HandleFunc("GET /{hash}", handler.GoTo())
	router.Handle("GET /link", middleware.IsAuthed(handler.GetAll(), deps.Config))
	router.Handle("GET /", handler.GetHomePage()) // mb Handle

}
func (h *HomeHandler) GetAll() http.Handler {
	panic("unimplemented")
}

func (h *HomeHandler) GoTo() func(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
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
