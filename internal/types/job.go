package types

import "time"

type Job struct {
	ID          string
	Command     string
	State       string
	Attempts    int
	MaxRetries  int
	BaseBackoff int
	NextRetryAt *time.Time
	LockedAt    *time.Time
	Output      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
