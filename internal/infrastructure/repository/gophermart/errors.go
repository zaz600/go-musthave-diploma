package gophermart

import "errors"

var (
	ErrUserNotFound = errors.New("userservice not found")
	ErrUserExists   = errors.New("userservice already exists")
)
