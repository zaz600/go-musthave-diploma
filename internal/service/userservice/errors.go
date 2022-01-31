package userservice

import "errors"

var (
	ErrUserExists = errors.New("userservice already exists")
	ErrAuth       = errors.New("invalid login or password")
)
