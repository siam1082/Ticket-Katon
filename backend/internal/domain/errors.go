package domain

import "errors"

var (
	ErrNotFound     = errors.New("record not found")
	ErrConflict     = errors.New("resource already exists")
	ErrUnauthorized = errors.New("unauthorized request")
	ErrForbidden    = errors.New("access forbidden")
	ErrSeatLocked   = errors.New("seat is already held or booked")
)