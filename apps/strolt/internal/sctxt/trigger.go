package sctxt

type TriggerType string

const (
	TSchedule TriggerType = "SCHEDULE"
	TApi      TriggerType = "API"
	TManual   TriggerType = "MANUAL"
)
