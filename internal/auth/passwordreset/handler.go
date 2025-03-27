package passwordreset

import (
	"encoding/json"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
)

type ResetHandlerDeps struct {
	*configs.Config
	*ResetService
}

type ResetHandler struct {
	*configs.Config
	*ResetService
}

func NewResetHandler(router *http.ServeMux, deps ResetHandlerDeps) {
	handler := &ResetHandler{
		Config:        deps.Config,
		ResetService:   deps.ResetService,
	}
	router.HandleFunc("POST /auth/reset", handler.Reset())
	router.HandleFunc("POST /auth/reset/verify", handler.VerifyCode())
}

func (h *ResetHandler) Reset() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
    var req RequestResetRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error("❌ error decoding request body: " + err.Error())
        http.Error(w, "Неверные данные", http.StatusBadRequest)
        return
    }
    logger.Info("📧 Сброс пароля для email: " + req.Email)
    if err := h.RequestReset(req.Email); err != nil {
        logger.Error("❌ error during password reset request: " + err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    logger.Info("✅ Сброс пароля успешно инициирован для email: " + req.Email)
    w.WriteHeader(http.StatusOK)
    }
}

func (h *ResetHandler) VerifyCode() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
    var req VerifyCodeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Неверные данные", http.StatusBadRequest)
        return
    }
    if err := h.VerifyCodeByEmail(req.Email, req.Code); err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }
    w.WriteHeader(http.StatusOK)
    }
}

