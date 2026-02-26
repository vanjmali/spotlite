package types

type EntityStatus string

const (
	StatusDeletionInProgress EntityStatus = "DELETION_IN_PROGRESS"
	StatusActive             EntityStatus = "ACTIVE"
)
