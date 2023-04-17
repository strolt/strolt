package sctxt

type OperationType string // @name OperationType

const (
	OpTypeBackup    OperationType = "BACKUP"
	OpTypePrune     OperationType = "PRUNE"
	OpTypeRestore   OperationType = "RESTORE"
	OpTypeSnapshots OperationType = "SNAPSHOTS"
	OpTypeStats     OperationType = "STATS"
)
