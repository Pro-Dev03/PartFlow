package audit

import "errors"

var (
	// ErrAuditLogNotFound is returned when audit log is not found
	ErrAuditLogNotFound = errors.New("audit log not found")

	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidAction is returned when action is invalid
	ErrInvalidAction = errors.New("invalid action")

	// ErrInvalidStatus is returned when status is invalid
	ErrInvalidStatus = errors.New("invalid status")

	// ErrInvalidEntityType is returned when entity type is invalid
	ErrInvalidEntityType = errors.New("invalid entity type")

	// ErrInvalidChanges is returned when changes format is invalid
	ErrInvalidChanges = errors.New("invalid changes format")
)
