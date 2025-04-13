package passwordreset

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/di"
	"github.com/ShopOnGO/ShopOnGO/pkg/email"
	"github.com/ShopOnGO/ShopOnGO/pkg/email/smtp"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type ResetService struct {
    Conf                *configs.Config
    SMTP                *smtp.SMTPSender
    Storage             di.IRedisResetRepository
    UserRepository      di.IUserRepository
}

func NewResetService(conf *configs.Config, smtpSender *smtp.SMTPSender, storage di.IRedisResetRepository, user di.IUserRepository) *ResetService {
	return &ResetService{
        Conf:           conf,
		SMTP:           smtpSender,
		Storage:        storage,
        UserRepository: user,
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
    emailInput := email.SendEmailInput{
        To:      toEmail,
		Subject: "Восстановление пароля",
        Body: code,
	}

    if err := service.SMTP.Send(emailInput); err != nil {
        logger.Error("❌ ошибка при отправке email: " + err.Error())
        return fmt.Errorf("failed to send email: %v", err)
    }

    logger.Info("✅ Код для восстановления пароля отправлен на email: " + toEmail)

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

	emailInput := email.SendEmailInput{
		To:      toEmail,
		Subject: "Восстановление пароля (повторно)",
		Body:    fmt.Sprintf("Ваш новый код для сброса пароля: %s", code),
	}

	if err := service.SMTP.Send(emailInput); err != nil {
		logger.Error("❌ ошибка при отправке email для " + toEmail + ": " + err.Error())
		return fmt.Errorf("failed to send email: %v", err)
	}

	logger.Info("✅ Код для восстановления пароля отправлен повторно на email: " + toEmail)
	return nil
}
