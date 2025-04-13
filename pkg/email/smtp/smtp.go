package smtp

import (
	"fmt"

	"github.com/ShopOnGO/ShopOnGO/pkg/email"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/go-gomail/gomail"
)

type SMTPSender struct {
	name string
	from string
	pass string
	host string
	port int
}

func NewSMTPSender(name,from, pass, host string, port int) *SMTPSender {
	return &SMTPSender{name: name, from: from, pass: pass, host: host, port: port}
}

func (s *SMTPSender) Send(input email.SendEmailInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", s.from)
	msg.SetHeader("To", input.To)
	msg.SetHeader("Subject", input.Subject)
	msg.SetBody("text/html", input.Body)

	dialer := gomail.NewDialer(s.host, s.port, s.name, s.pass)

	logger.Info(fmt.Sprintf("Попытка отправки email на %s через SMTP сервер %s:%d", input.To, s.host, s.port))

	if err := dialer.DialAndSend(msg); err != nil {
		logger.Error(fmt.Sprintf("❌ Ошибка при отправке email на %s: %s", input.To, err.Error()))
		return fmt.Errorf("failed to send email via smtp: %v", err)
	}

	logger.Info(fmt.Sprintf("✅ Email успешно отправлен на %s", input.To))

return nil
}