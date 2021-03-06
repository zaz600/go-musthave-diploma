package accrual

import (
	"errors"
)

var ErrWrongStatusCode = errors.New("wrong http status code")
var ErrUnknownAccrualStatus = errors.New("unknown accrual status")
var ErrFatalError = errors.New("fatal error")
var ErrTooManyRedirects = errors.New("too many requests")

type TooManyRequestsError struct {
	err           error
	RetryAfterSec int
}

func (e TooManyRequestsError) Error() string {
	return e.err.Error()
}

func (e TooManyRequestsError) Unwrap() error {
	return e.err
}

func NewTooManyRequestsError(retryAfterSec int) *TooManyRequestsError {
	return &TooManyRequestsError{
		RetryAfterSec: retryAfterSec,
		err:           ErrTooManyRedirects,
	}
}
