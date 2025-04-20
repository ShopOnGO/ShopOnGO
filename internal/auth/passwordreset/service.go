package passwordreset

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/di"
	"github.com/ShopOnGO/ShopOnGO/pkg/kafkaService"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type ResetService struct {
	Conf           *configs.Config
	Kafka          *kafkaService.KafkaService
	Storage        di.IRedisResetRepository
	UserRepository di.IUserRepository
}

func NewResetService(conf *configs.Config, storage di.IRedisResetRepository, user di.IUserRepository, kafka *kafkaService.KafkaService) *ResetService {
	return &ResetService{
		Conf:           conf,
		Storage:        storage,
		UserRepository: user,
		Kafka:          kafka,
	}
}

func GenerateCode() (string, error) {
	var num int64
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	num = (int64(b[0])<<24 | int64(b[1])<<16 | int64(b[2])<<8 | int64(b[3])) % 1000000

	return fmt.Sprintf("%06d", num), nil
}

func (service *ResetService) RequestReset(toEmail string) error {
	logger.Info("🔐 Запрос на сброс пароля для email: " + toEmail)

	user, err := service.UserRepository.FindByEmail(toEmail)
	if err != nil {
		logger.Error("❌ ошибка при поиске пользователя по email: " + err.Error())
		return err
	}

	if user.Provider == "google" {
		logger.Error("❌ ошибка сброс пароля для зарегистрированного через Google пользователя")
		return errors.New("сброс пароля недоступен для пользователей, зарегистрированных через Google")
	}

	requests, err := service.Storage.GetResetCodeCount(toEmail)
	if err != nil {
		logger.Error("❌ ошибка при получении количества запросов для email: " + err.Error())
		return err
	}
	if requests >= service.Conf.Code.MaxRequests {
		return errors.New("превышено количество запросов на сброс пароля, попробуйте позже")
	}

	// Увеличиваем счетчик запросов
	if err := service.Storage.IncrementResetCodeCount(toEmail, service.Conf.Code.RateLimitTTL); err != nil {
		logger.Error("❌ ошибка при обновлении количества запросов для email: " + err.Error())
		return err
	}

	code, err := GenerateCode()
	if err != nil {
		logger.Error("❌ ошибка при генерации кода: " + err.Error())
		return err
	}
	expiresAt := time.Now().Add(service.Conf.Code.CodeTTL)
	if err := service.Storage.SaveToken(toEmail, code, expiresAt); err != nil {
		logger.Error("❌ ошибка при сохранении токена в хранилище: " + err.Error())
		return err
	}
	// emailInput := email.SendEmailInput{
	//     To:      toEmail,
	// 	Subject: "Восстановление пароля",
	//     Body: code,
	// }

	// if err := service.SMTP.Send(emailInput); err != nil {
	//     logger.Error("❌ ошибка при отправке email: " + err.Error())
	//     return fmt.Errorf("failed to send email: %v", err)
	// }

	// logger.Info("✅ Код для восстановления пароля отправлен на email: " + toEmail)

	// return nil
	// 🔥 Подготовка Kafka-события
	event := map[string]interface{}{
		"action":   "create",
		"category": "AUTHRESET",
		"subtype":  "SEND_RESET_CODE",
		"userID":   0, // Email теперь не в userID, а в Payload
		"wasInDlq": false,
		"payload": map[string]interface{}{
			"code":      code,
			"subject":   "Восстановление пароля",
			"expiresAt": expiresAt.Unix(),
			"email":     toEmail, // <-- добавлен email
		},
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.Error("❌ ошибка сериализации события: " + err.Error())
		return err
	}

	key := []byte(fmt.Sprintf("reset-%s", toEmail))
	if err := service.Kafka.Produce(context.Background(), key, eventBytes); err != nil {
		logger.Errorf("❌ ошибка отправки сообщения в Kafka: %v", err)
		return err
	}

	logger.Info("📨 Событие сброса пароля отправлено в Kafka для email: " + toEmail)
	return nil
}

func (service *ResetService) VerifyCodeByEmail(toEmail, code string) error {
	storedCode, expiresAt, err := service.Storage.GetToken(toEmail)
	if err != nil {
		return err
	}
	if time.Now().After(expiresAt) {
		return errors.New("код истек")
	}
	if storedCode != code {
		return errors.New("неверный код")
	}
	// надо ли
	// return service.Storage.DeleteToken(toEmail)
	return nil
}

func (service *ResetService) ResetPassword(toEmail, newPassword string) error {
	user, err := service.UserRepository.FindByEmail(toEmail)
	if err != nil {
		logger.Error("❌ ошибка при поиске пользователя по email: " + err.Error())
		return fmt.Errorf("пользователь не найден")
	}

	if user.Provider == "google" {
		return errors.New("сброс пароля недоступен для пользователей, зарегистрированных через Google")
	}

	// storedCode, expiresAt, err := service.Storage.GetToken(toEmail)
	_, expiresAt, err := service.Storage.GetToken(toEmail)
	if err != nil {
		logger.Error("❌ не удалось получить токен для email " + toEmail + ": " + err.Error())
		return fmt.Errorf("код не найден, запросите сброс пароля повторно")
	}
	if time.Now().After(expiresAt) {
		return errors.New("код истек, запросите новый")
	}
	// if storedCode != code {
	// 	return errors.New("неверный код")
	// }
	if err := service.Storage.DeleteToken(toEmail); err != nil {
		logger.Error("❌ ошибка при удалении токена для email " + toEmail + ": " + err.Error())
		return err
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)

	if err != nil {
		return fmt.Errorf("FailedToHashNewPassword: %w", err)
	}

	logger.Info("Хешированный пароль:", string(newPasswordHash))
	if err := service.UserRepository.UpdateUserPassword(user.ID, string(newPasswordHash)); err != nil {
		logger.Error("❌ ошибка при обновлении пароля для email " + toEmail + ": " + err.Error())
		return err
	}
	logger.Info("✅ Пароль успешно обновлен для email: " + toEmail)
	return nil
}

func (service *ResetService) ResendCode(toEmail string) error {
	user, err := service.UserRepository.FindByEmail(toEmail)
	if err != nil {
		logger.Error("❌ ошибка при поиске пользователя по email: " + err.Error())
		return fmt.Errorf("пользователь не найден")
	}

	if user.Provider == "google" {
		return errors.New("сброс пароля недоступен для пользователей, зарегистрированных через Google")
	}

	requests, err := service.Storage.GetResetCodeCount(toEmail)
	if err != nil {
		logger.Error("❌ ошибка при получении количества запросов для email: " + err.Error())
		return err
	}
	if requests >= service.Conf.Code.MaxRequests {
		return errors.New("превышено количество запросов на сброс пароля, попробуйте позже")
	}
	if err := service.Storage.IncrementResetCodeCount(toEmail, service.Conf.Code.RateLimitTTL); err != nil {
		logger.Error("❌ ошибка при обновлении счетчика запросов для email: " + err.Error())
		return err
	}

	// Генерируем новый код
	code, err := GenerateCode()
	if err != nil {
		logger.Error("❌ ошибка при генерации кода: " + err.Error())
		return err
	}
	expiresAt := time.Now().Add(service.Conf.Code.CodeTTL)

	if err := service.Storage.SaveToken(toEmail, code, expiresAt); err != nil {
		logger.Error("❌ ошибка при сохранении токена для email " + toEmail + ": " + err.Error())
		return err
	}

	// emailInput := email.SendEmailInput{
	// 	To:      toEmail,
	// 	Subject: "Восстановление пароля (повторно)",
	// 	Body:    fmt.Sprintf("Ваш новый код для сброса пароля: %s", code),
	// }

	// if err := service.SMTP.Send(emailInput); err != nil {
	// 	logger.Error("❌ ошибка при отправке email для " + toEmail + ": " + err.Error())
	// 	return fmt.Errorf("failed to send email: %v", err)
	// }

	// logger.Info("✅ Код для восстановления пароля отправлен повторно на email: " + toEmail)
	// return nil
	event := map[string]interface{}{
		"action":   "create",
		"category": "AUTHRESET",
		"subtype":  "RESET_CODE",
		"userID":   0,
		"wasInDlq": false,
		"payload": map[string]interface{}{
			"code":      code,
			"subject":   "Восстановление пароля (повторно)",
			"expiresAt": expiresAt.Unix(),
			"email":     toEmail, // 💡 обязательно!
		},
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.Error("❌ ошибка сериализации события: " + err.Error())
		return err
	}

	key := []byte(fmt.Sprintf("reset-resend-%s", toEmail))
	if err := service.Kafka.Produce(context.Background(), key, eventBytes); err != nil {
		logger.Errorf("❌ ошибка отправки сообщения в Kafka: %v", err)
		return err
	}

	logger.Info("📨 Повторное событие восстановления пароля отправлено в Kafka для email: " + toEmail)
	return nil
}
