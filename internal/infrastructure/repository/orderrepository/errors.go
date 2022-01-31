package orderrepository

import "errors"

var ErrOrderExists = errors.New("order already exists")
var ErrOrderOwnedByAnotherUser = errors.New("order uploaded by another user")
