package types

type SongStatus string

const (
	StatusDeletionInProgress SongStatus = "DELETION_IN_PROGRESS"
	StatusActive             SongStatus = "ACTIVE"
)
