package ports

import "errors"

type NonRetriableError struct {
	Err error
}

func (e *NonRetriableError) Error() string {
	return e.Err.Error()
}

func (e *NonRetriableError) Unwrap() error {
	return e.Err
}

func NewNonRetriableError(err error) error {
	return &NonRetriableError{Err: err}
}

func IsNonRetriable(err error) bool {
	var nonRetriable *NonRetriableError
	return errors.As(err, &nonRetriable)
}
