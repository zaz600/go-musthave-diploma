package sessionrepository

import "errors"

var ErrSessionExists = errors.New("session already exists")
var ErrSessionNotFound = errors.New("session not found")
