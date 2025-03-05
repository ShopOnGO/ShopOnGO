package req

import "github.com/go-playground/validator"

func Validate[T any](payload T) error {
	validate := validator.New()
	err := validate.Struct(payload)
	return err
}
