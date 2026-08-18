package notifications

import "errors"

var (
	// ErrNotificationNotFound is returned when notification is not found
	ErrNotificationNotFound = errors.New("notification not found")

	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidNotificationType is returned when notification type is invalid
	ErrInvalidNotificationType = errors.New("invalid notification type")

	// ErrInvalidNotificationStatus is returned when notification status is invalid
	ErrInvalidNotificationStatus = errors.New("invalid notification status")

	// ErrInvalidPriority is returned when priority is invalid
	ErrInvalidPriority = errors.New("invalid priority")

	// ErrNotificationExpired is returned when notification is expired
	ErrNotificationExpired = errors.New("notification expired")

	// ErrNotificationAlreadyRead is returned when notification is already read
	ErrNotificationAlreadyRead = errors.New("notification already read")

	// ErrPreferencesNotFound is returned when notification preferences are not found
	ErrPreferencesNotFound = errors.New("notification preferences not found")
)
