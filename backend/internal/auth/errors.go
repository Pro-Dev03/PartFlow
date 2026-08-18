package auth

import "errors"

var (
	// ErrInvalidCredentials is returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrUserExists is returned when user already exists
	ErrUserExists = errors.New("user already exists")

	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired is returned when token is expired
	ErrTokenExpired = errors.New("token expired")

	// ErrInvalidPassword is returned when password is invalid
	ErrInvalidPassword = errors.New("invalid password")

	// ErrInactiveUser is returned when user is inactive
	ErrInactiveUser = errors.New("user is inactive")

	// ErrUnauthorized is returned when user is not authorized
	ErrUnauthorized = errors.New("unauthorized")
)
