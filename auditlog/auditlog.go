package auditlog

import "time"

type Service interface {
	Create(int64, string) error
	List(ListFilter) ([]AuditLog, int, error)
}

type AuditLog struct {
	UserId       string    `db:"user_id"`
	UserFullName string    `db:"user_full_name"`
	RecordedAt   time.Time `db:"recorded_at"`
	Description  string    `db:"description"`
	Count        int       `db:"count"`
}
