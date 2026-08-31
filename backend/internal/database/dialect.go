package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// IsSQLite reports whether the active connection uses the local SQLite store.
func IsSQLite(db *sqlx.DB) bool {
	return db != nil && strings.EqualFold(db.DriverName(), "sqlite")
}

// NowSQL returns the current timestamp expression for the active database.
func NowSQL(db *sqlx.DB) string {
	if IsSQLite(db) {
		return "CURRENT_TIMESTAMP"
	}
	return "NOW()"
}

// Placeholder returns a database-specific positional parameter.
func Placeholder(db *sqlx.DB, position int) string {
	if IsSQLite(db) {
		return "?"
	}
	return "$" + fmt.Sprint(position)
}

// ParseTimestamp converts SQLite TEXT timestamps and PostgreSQL timestamptz values into time.Time.
func ParseTimestamp(value any) (time.Time, error) {
	switch v := value.(type) {
	case nil:
		return time.Time{}, nil
	case time.Time:
		return v, nil
	case string:
		for _, layout := range []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05-07:00",
			"2006-01-02 15:04:05",
			"2006-01",
			"2006-01-02",
		} {
			if parsed, err := time.Parse(layout, v); err == nil {
				return parsed, nil
			}
		}
		return time.Time{}, fmt.Errorf("unsupported timestamp format: %q", v)
	case []byte:
		return ParseTimestamp(string(v))
	default:
		return time.Time{}, fmt.Errorf("unsupported timestamp type %T", value)
	}
}
