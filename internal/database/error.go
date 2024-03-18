package database

import (
	"fmt"
)

type ConflictError struct {
	Err error
}

func NewErrorConflict(err error) error {
	return &ConflictError{Err: err}
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%v : %v", e.Err, "already exists")
}
