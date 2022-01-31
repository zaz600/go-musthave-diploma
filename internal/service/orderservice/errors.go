package orderservice

import "errors"

var ErrOrderExists = errors.New("order already exists")
var ErrOrderOwnedByAnotherUser = errors.New("order uploaded by another user")
var ErrInvalidOrderFormat = errors.New("order format error")
