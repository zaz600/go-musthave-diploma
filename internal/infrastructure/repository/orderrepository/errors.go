package orderrepository

import "errors"

var ErrOrderExists = errors.New("order already exists")
var ErrOrderNotFound = errors.New("order not found")
var ErrOrderOwnedByAnotherUser = errors.New("order uploaded by another user")
