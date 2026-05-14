package enums

type SessionStatus string

const (
	SessionStatusPending  SessionStatus = "pending"
	SessionStatusActive   SessionStatus = "active"
	SessionStatusInactive SessionStatus = "inactive"
	SessionStatusFailed   SessionStatus = "failed"
)
