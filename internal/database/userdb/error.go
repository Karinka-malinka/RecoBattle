package userdb

import (
	"fmt"
)

type ErrConflict struct {
	Err error
}

func NewErrorConflict(err error) error {
	return &ErrConflict{Err: err}
}

func (e *ErrConflict) Error() string {
	return fmt.Sprintf("%v : %v", e.Err, "already exists")
}
