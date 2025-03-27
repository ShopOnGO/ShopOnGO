package passwordreset

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/di"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/email"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/email/smtp"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type ResetService struct {
    Conf                *configs.Config
    SMTP                *smtp.SMTPSender
    Storage             di.IURedisResetRepository
    UserRepository      di.IUserRepository
}

func NewResetService(conf *configs.Config, smtpSender *smtp.SMTPSender, storage di.IURedisResetRepository, user di.IUserRepository) *ResetService {
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
    
    _, err := service.UserRepository.FindByEmail(toEmail)
    if err != nil {
        logger.Error("❌ ошибка при поиске пользователя по email: " + err.Error())        
        return nil // или return стандартную заглушку
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

func (service *ResetService) ResetPassword(code, toEmail, newPassword string) error {
    storedCode, expiresAt, err := service.Storage.GetToken(toEmail)
	if err != nil {
		logger.Error("❌ не удалось получить токен для email " + toEmail + ": " + err.Error())
		return fmt.Errorf("код не найден, запросите сброс пароля повторно")
	}
	if time.Now().After(expiresAt) {
		return errors.New("код истек, запросите новый")
	}
	if storedCode != code {
		return errors.New("неверный код")
	}
	if err := service.Storage.DeleteToken(toEmail); err != nil {
		logger.Error("❌ ошибка при удалении токена для email " + toEmail + ": " + err.Error())
		return err
	}

	user, err := service.UserRepository.FindByEmail(toEmail)
	if err != nil {
		logger.Error("❌ ошибка при поиске пользователя по email: " + err.Error())
		return fmt.Errorf("пользователь не найден")
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
	_, err := service.UserRepository.FindByEmail(toEmail)
	if err != nil {
		logger.Error("❌ ошибка при поиске пользователя по email: " + err.Error())
		return fmt.Errorf("пользователь не найден")
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
