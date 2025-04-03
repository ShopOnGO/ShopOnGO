package email

import (
	"bytes"
	"errors"
	"html/template"

)

type SendEmailInput struct {
	To      string
	Subject string
	Body    string
}

type Sender interface {
	Send(input SendEmailInput) error
}

func (e *SendEmailInput) GenerateBodyFromHTML(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	e.Body = buf.String()

	return nil
}

func (e *SendEmailInput) Validate() error {
	if e.To == "" {
		return errors.New("empty to")
	}

	if e.Subject == "" {
		return errors.New("empty subject")
	}

	if e.Body == "" {
		return errors.New("empty body")
	}

	return nil
}