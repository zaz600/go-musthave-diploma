package withdrawalrepository

import "errors"

var ErrWithdrawalExists = errors.New("withdrawal already exists")
var ErrWithdrawalOwnedByAnotherUser = errors.New("withdrawal uploaded by another user")
