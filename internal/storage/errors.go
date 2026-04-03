package storage

import "errors"

// ErrUserAlreadyExists is returned when attempting to create a user with a duplicate email.
var ErrUserAlreadyExists = errors.New("user already exists")

// ErrUserNotFound is returned when no user matches the query.
var ErrUserNotFound = errors.New("user not found")

// ErrSecretNotFound is returned when no secret matches the query.
var ErrSecretNotFound = errors.New("secret not found")
