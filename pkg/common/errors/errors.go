// pkg/common/errors/errors.go
package errors

import (
	"errors"
	"fmt"
)

type CustomError struct {
	Code int
	Err  error
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("code=%d, error=%v", e.Code, e.Err)
}

func New(code int, msg string) error {
	return &CustomError{code, errors.New(msg)}
}
