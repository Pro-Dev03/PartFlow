package business

import "strings"

var OpenDebtStatuses = []string{"pending", "partial", "overdue"}

func IsOpenDebtStatus(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	for _, openStatus := range OpenDebtStatuses {
		if status == openStatus {
			return true
		}
	}
	return false
}

func OpenDebtStatusSQL(column string) string {
	quoted := make([]string, len(OpenDebtStatuses))
	for index, status := range OpenDebtStatuses {
		quoted[index] = "'" + status + "'"
	}
	return column + " IN (" + strings.Join(quoted, ", ") + ")"
}

func IsDebtOverdue(remainingAmount float64, dueDateIsBeforeToday bool, status string) bool {
	return remainingAmount > 0 && dueDateIsBeforeToday && IsOpenDebtStatus(status)
}
