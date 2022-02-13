package gophermartservice

import "errors"

var (
	ErrUserExists              = errors.New("user already exists")
	ErrAuth                    = errors.New("invalid login or password")
	ErrOrderExists             = errors.New("order already exists")
	ErrOrderOwnedByAnotherUser = errors.New("order uploaded by another user")
	ErrInvalidOrderFormat      = errors.New("order format error")
)
