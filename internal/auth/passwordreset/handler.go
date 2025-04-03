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
    router.HandleFunc("POST /auth/reset/password", handler.ResetPassword())
    router.HandleFunc("POST /auth/reset/resend", handler.ResendCode())
}

// Reset инициирует процедуру сброса пароля, генерирует код и отправляет его на указанный email
// @Summary      Запрос на сброс пароля
// @Description  Генерирует код сброса пароля и отправляет его на email пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  ResetRequest  true  "Данные для запроса сброса пароля"
// @Success      200   {string}  string  "Сброс пароля успешно инициирован"
// @Failure      400   {string}  string  "Неверные данные"
// @Failure      500   {string}  string  "Ошибка сервера при инициировании сброса пароля"
// @Router       /auth/reset [post]
func (h *ResetHandler) Reset() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		var req ResetRequest
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

		response := map[string]string{"message": "Password reset request initiated successfully"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
    }
}

// VerifyCode проверяет корректность кода, отправленного пользователю для сброса пароля
// @Summary      Верификация кода сброса пароля
// @Description  Проверяет, соответствует ли указанный код сохраненному для email, и подтверждает запрос сброса пароля
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  VerifyCodeRequest  true  "Данные для верификации кода"
// @Success      200   {string}  string  "Код подтвержден"
// @Failure      400   {string}  string  "Неверные данные"
// @Failure      401   {string}  string  "Неверный или просроченный код"
// @Router       /auth/reset/verify [post]
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
		response := map[string]string{"message": "Verification code is valid"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
    }
}

// ResetPassword обновляет пароль пользователя после проверки кода сброса пароля.
// @Summary      Обновление пароля
// @Description  Проверяет предоставленный код сброса и, при корректном совпадении, обновляет пароль пользователя.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  ResetPasswordRequest  true  "Данные для обновления пароля"
// @Success      200   {string} string  "Пароль успешно обновлен"
// @Failure      400   {string} string  "Неверные данные"
// @Failure      401   {string} string  "Неверный или просроченный код"
// @Failure      500   {string} string  "Ошибка сервера при обновлении пароля"
// @Router       /auth/reset/password [post]
func (h *ResetHandler) ResetPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ResetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("❌ error decoding request body: " + err.Error())
			http.Error(w, "Неверные данные", http.StatusBadRequest)
			return
		}
		logger.Info("🔑 Запрос на установку нового пароля для email: " + req.Email)
		if err := h.ResetService.ResetPassword(req.Email, req.NewPassword); err != nil {
			logger.Error("❌ ошибка при установке нового пароля для email " + req.Email + ": " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Info("✅ Пароль успешно обновлен для email: " + req.Email)
		response := map[string]string{"message": "Password successfully updated"}
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
	}
}

// ResendCode генерирует и отправляет повторно код для сброса пароля пользователю.
// @Summary      Повторная отправка кода сброса пароля
// @Description  Генерирует новый код сброса пароля и отправляет его на указанный email пользователя.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  ResetRequest  true  "Данные для запроса повторной отправки кода"
// @Success      200   {string} string  "Код успешно отправлен повторно"
// @Failure      400   {string} string  "Неверные данные"
// @Failure      500   {string} string  "Ошибка сервера при отправке кода"
// @Router       /auth/reset/resend [post]
func (h *ResetHandler) ResendCode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        var req ResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("❌ error decoding request body: " + err.Error())
			http.Error(w, "Неверные данные", http.StatusBadRequest)
			return
		}
		logger.Info("📧 Повторная отправка кода для email: " + req.Email)
		if err := h.ResetService.ResendCode(req.Email); err != nil {
			logger.Error("❌ ошибка при повторной отправке кода для email " + req.Email + ": " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Info("✅ Код успешно отправлен повторно для email: " + req.Email)
		response := map[string]string{"message": "Reset code resent successfully"}
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
	}
}
