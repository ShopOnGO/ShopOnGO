package passwordreset

import (
	"encoding/json"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/gorilla/mux"
)

type ResetHandlerDeps struct {
	*configs.Config
	*ResetService
}

type ResetHandler struct {
	*configs.Config
	*ResetService
}

func NewResetHandler(router *mux.Router, deps ResetHandlerDeps) {
	handler := &ResetHandler{
		Config:       deps.Config,
		ResetService: deps.ResetService,
	}
	router.Handle("/auth/reset", handler.Reset()).Methods("POST")
	router.Handle("/auth/reset/verify", handler.VerifyCode()).Methods("POST")
	router.Handle("/auth/reset/password", handler.ResetPassword()).Methods("POST")
	router.Handle("/auth/reset/resend", handler.ResendCode()).Methods("POST")
}

// Reset initiates the password reset process.
// It generates a reset code and sends it to the provided email address.
// @Summary      Request Password Reset
// @Description  Generates a password reset code and sends it to the user's email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  ResetRequest  true  "Data for password reset request"
// @Success      200   {string}  string  "Password reset successfully initiated"
// @Failure      400   {string}  string  "Invalid input data"
// @Failure      500   {string}  string  "Server error during password reset initiation"
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

// VerifyCode checks the validity of the reset code sent to the user.
// @Summary      Verify Reset Code
// @Description  Validates the provided reset code against the one stored for the email and confirms the reset request.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  VerifyCodeRequest  true  "Data for code verification"
// @Success      200   {string}  string  "Code verified successfully"
// @Failure      400   {string}  string  "Invalid input data"
// @Failure      401   {string}  string  "Invalid or expired code"
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

// ResetPassword updates the user's password after successful code verification.
// @Summary      Update Password
// @Description  Verifies the provided reset code and, if valid, updates the user's password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  ResetPasswordRequest  true  "Data for password update"
// @Success      200   {string} string  "Password successfully updated"
// @Failure      400   {string} string  "Invalid input data"
// @Failure      401   {string} string  "Invalid or expired code"
// @Failure      500   {string} string  "Server error during password update"
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

// ResendCode generates and resends the password reset code to the user.
// @Summary      Resend Reset Code
// @Description  Generates a new password reset code and sends it to the user's email address.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  ResetRequest  true  "Data for resending code request"
// @Success      200   {string} string  "Code successfully resent"
// @Failure      400   {string} string  "Invalid input data"
// @Failure      500   {string} string  "Server error during code resend"
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
