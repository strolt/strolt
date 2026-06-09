package sctxt

// TriggerType identifies what initiated an operation.
type TriggerType string //	@name	TriggerType

// Supported trigger types.
const (
	TSchedule TriggerType = "SCHEDULE"
	TApi      TriggerType = "API"
	TManual   TriggerType = "MANUAL"
)
