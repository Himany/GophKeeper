package storage

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")

	ErrUserExists = errors.New("user already exists")

	ErrEntryNotFound = errors.New("entry not found")

	ErrEntryExists = errors.New("entry already exists")

	ErrInvalidCredentials = errors.New("invalid credentials")
)
