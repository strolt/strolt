package sctxt

// OperationType identifies the kind of operation performed by a task.
type OperationType string //	@name	OperationType

// Supported operation types.
const (
	OpTypeBackup    OperationType = "BACKUP"
	OpTypePrune     OperationType = "PRUNE"
	OpTypeRestore   OperationType = "RESTORE"
	OpTypeSnapshots OperationType = "SNAPSHOTS"
	OpTypeStats     OperationType = "STATS"
)
