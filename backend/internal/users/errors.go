package users

import "errors"

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrUserEmailExists is returned when a user email already exists
	ErrUserEmailExists = errors.New("user email already exists")

	// ErrInvalidUserData is returned when user data is invalid
	ErrInvalidUserData = errors.New("invalid user data")

	// ErrInvalidPassword is returned when password is invalid
	ErrInvalidPassword = errors.New("invalid password")

	// ErrCurrentPasswordIncorrect is returned when current password is incorrect
	ErrCurrentPasswordIncorrect = errors.New("current password is incorrect")

	// ErrPasswordTooShort is returned when password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
)
