package link

import (
	"net/http"
	"strconv"

	"github.com/ShopOnGO/ShopOnGO/configs"
	_ "github.com/ShopOnGO/ShopOnGO/docs"
	"github.com/ShopOnGO/ShopOnGO/pkg/event"
	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/pkg/req"
	"github.com/ShopOnGO/ShopOnGO/pkg/res"
	"github.com/gorilla/mux"
)

type LinkHandlerDeps struct { // содержит все необходимые элементы заполнения. это DC
	LinkService    *LinkService
	LinkRepository *LinkRepository
	Config         *configs.Config
	EventBus       *event.EventBus
}
type LinkHandler struct { // это уже рабоая структура
	LinkService    *LinkService
	LinkRepository *LinkRepository
	EventBus       *event.EventBus
}

func NewLinkHandler(router *mux.Router, deps LinkHandlerDeps) {
	handler := &LinkHandler{
		LinkRepository: deps.LinkRepository,
		LinkService:    deps.LinkService,
		EventBus:       deps.EventBus,
	}
	router.Handle("/link", middleware.IsAuthed(handler.Create(), deps.Config)).Methods("POST")
	router.Handle("/link/{id}", middleware.IsAuthed(handler.Update(), deps.Config)).Methods("PATCH")
	router.Handle("/link/{id}", middleware.IsAuthed(handler.Delete(), deps.Config)).Methods("DELETE")
	router.HandleFunc("/goto/{hash}", handler.GoTo()).Methods("GET")
	router.Handle("/link", middleware.IsAuthed(handler.GetAll(), deps.Config)).Methods("GET")	
}

// Create создает новую короткую ссылку
// @Summary        Создание короткой ссылки
// @Description    Генерирует короткую ссылку по переданному URL и сохраняет ее в базе
// @Tags          link
// @Accept        json
// @Produce       json
// @Security      ApiKeyAuth
// @Param         link body LinkCreateRequest true "Данные для создания ссылки"
// @Success       201 {object} Link
// @Failure       400 {string} string "Некорректный запрос"
// @Router        /link [post]
func (h *LinkHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := req.HandleBody[LinkCreateRequest](&w, r)
		if err != nil { // валидация слабая, только на http:
			return
		}

		createdLink, err := h.LinkService.CreateLink(body.Url) // может быть коллизия
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res.Json(w, createdLink, 201)
	}
}

// Update обновляет существующую короткую ссылку
// @Summary        Обновление ссылки
// @Description    Изменяет URL или хеш существующей короткой ссылки
// @Tags          link
// @Accept        json
// @Produce       json
// @Security      ApiKeyAuth
// @Param         id path int true "ID ссылки"
// @Param         link body LinkUpdateRequest true "Данные для обновления ссылки"
// @Success       200 {object} Link
// @Failure       400 {string} string "Некорректный запрос"
// @Failure       404 {string} string "Ссылка не найдена"
// @Router        /link/{id} [put]
func (h *LinkHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := req.HandleBody[LinkUpdateRequest](&w, r)
		if err != nil {
			return
		}
		idString := r.PathValue("id")
		id, err := strconv.ParseUint(idString, 10, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		link, err := h.LinkService.UpdateLink(uint(id), body.Url, body.Hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res.Json(w, link, 201)

	}
}

// Delete удаляет короткую ссылку по ID
// @Summary        Удаление ссылки
// @Description    Удаляет существующую короткую ссылку из базы данных
// @Tags          link
// @Security      ApiKeyAuth
// @Param         id path int true "ID ссылки"
// @Success       200 {string} string "Ссылка успешно удалена"
// @Failure       400 {string} string "Некорректный ID"
// @Failure       404 {string} string "Ссылка не найдена"
// @Failure       500 {string} string "Ошибка сервера"
// @Router        /link/{id} [delete]
func (h *LinkHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idString := r.PathValue("id")
		id, err := strconv.ParseUint(idString, 10, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = h.LinkService.DeleteLink(uint(id))
		if err != nil {
			if err.Error() == "link not found" {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		res.Json(w, nil, 200)

	}
}

// GoTo перенаправляет пользователя на оригинальный URL по хешу ссылки
// @Summary        Редирект по хешу
// @Description    Ищет короткую ссылку в базе по хешу и выполняет перенаправление
// @Tags          link
// @Param         hash path string true "Хеш ссылки"
// @Success       307 {string} string "Перенаправление"
// @Failure       404 {string} string "Ссылка не найдена"
// @Router        /{hash} [get]
func (h *LinkHandler) GoTo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := r.PathValue("hash")

		link, err := h.LinkService.GoTo(hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Redirect(w, r, link.Url, http.StatusTemporaryRedirect)
	}
}

// GetAll возвращает список коротких ссылок с пагинацией
// @Summary        Получить все ссылки
// @Description    Возвращает список всех коротких ссылок с возможностью пагинации
// @Tags          link
// @Accept        json
// @Produce       json
// @Security      ApiKeyAuth
// @Param         limit  query int false "Количество ссылок (по умолчанию 10)"
// @Param         offset query int false "Смещение (по умолчанию 0)"
// @Success       200 {object} GetAllLinksResponse
// @Failure       400 {string} string "Некорректные параметры limit/offset"
// @Router        /link [get]
func (h *LinkHandler) GetAll() http.HandlerFunc { //recheck
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		} // here ->
		links, count, err := h.LinkService.GetAll(limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		linksResponse := GetAllLinksResponse{
			Links: links,
			Count: count,
		}

		res.Json(w, linksResponse, 200)
	}
}
