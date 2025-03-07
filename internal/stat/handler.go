package stat

import (
	"net/http"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	_ "github.com/ShopOnGO/ShopOnGO/prod/docs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/res"
)

const (
	GroupByDay   = "day"
	GroupByMonth = "month"
)

type StatHandlerDeps struct { // содержит все необходимые элементы заполнения. это DC
	StatRepository *StatRepository
	Config         *configs.Config
}
type StatHandler struct { // это уже рабоая структура
	StatRepository *StatRepository
}

func NewStatHandler(router *http.ServeMux, deps StatHandlerDeps) {
	handler := &StatHandler{
		StatRepository: deps.StatRepository,
	}
	router.Handle("GET /stat", middleware.IsAuthed(handler.GetStat(), deps.Config))

}

// GetStat получает статистику переходов по ссылкам за указанный период
// @Summary      Получить статистику переходов
// @Description  Возвращает агрегированную статистику по количеству переходов, сгруппированную по дням или месяцам
// @Tags         statistics
// @Accept       json
// @Produce      json
// @Param        from query string true  "Начальная дата (формат: YYYY-MM-DD)"
// @Param        to   query string true  "Конечная дата (формат: YYYY-MM-DD)"
// @Param        by   query string true  "Группировка (допустимые значения: day, month)"
// @Success      200  {object}  []GetStatResponse "Успешный ответ со статистикой"
// @Failure      400  {string}  string  "Некорректные параметры запроса"
// @Router       /stats [get]
func (h *StatHandler) GetStat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
		if err != nil {
			http.Error(w, "Invalid from param", http.StatusBadRequest)
			return
		}
		to, err := time.Parse("2006-01-02", r.URL.Query().Get("to")) // под auery parzms можно сделать отдельный валидатор чтобы не повторяться дважды
		if err != nil {
			http.Error(w, "Invalid to param", http.StatusBadRequest)
			return
		}
		by := r.URL.Query().Get("by")
		if by != GroupByDay && by != GroupByMonth {
			http.Error(w, "Invalid by param", http.StatusBadRequest)
			return
		}
		stats := h.StatRepository.GetStats(by, from, to)
		res.Json(w, stats, 200)
	}
}
