package errors

import "errors"

var (
	ErrNotFound     = errors.New("not_found")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidState = errors.New("invalid_state")
	ErrCapacity     = errors.New("capacity_exceeded")
	ErrExpired      = errors.New("expired")
)

type Coded struct {
	Code string
	Err  error
}

func (e *Coded) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *Coded) Unwrap() error { return e.Err }
func Wrap(code string, err error) error {
	if err == nil {
		return nil
	}
	return &Coded{Code: code, Err: err}
}
