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

func (service *ResetService) VerifyCodeByEmail(email, code string) error {
    storedCode, expiresAt, err := service.Storage.GetToken(email)
    if err != nil {
        return err
    }
    if time.Now().After(expiresAt) {
        return errors.New("код истек")
    }
    if storedCode != code {
        return errors.New("неверный код")
    }
    return service.Storage.DeleteToken(email)
}
