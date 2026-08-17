package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func Struct(value any) error {

	if err := validator.New().Struct(value); err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}

	return nil
}
