package sctxt

type OperationType string

const (
	OpTypeBackup    OperationType = "BACKUP"
	OpTypePrune     OperationType = "PRUNE"
	OpTypeRestore   OperationType = "RESTORE"
	OpTypeSnapshots OperationType = "SNAPSHOTS"
)
