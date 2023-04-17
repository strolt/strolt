package sctxt

type TriggerType string // @name TriggerType

const (
	TSchedule TriggerType = "SCHEDULE"
	TApi      TriggerType = "API"
	TManual   TriggerType = "MANUAL"
)
