package enums

type CheckinStatus string

const (
	CheckinStatusPending    CheckinStatus = "PENDING"
	CheckinStatusInProgress CheckinStatus = "IN_PROGRESS"
	CheckinStatusCompleted  CheckinStatus = "COMPLETED"
	CheckinStatusFailed     CheckinStatus = "FAILED"
	CheckinStatusMissed     CheckinStatus = "MISSED"
)

type MedicalStatus string

const (
	MedicalStatusNormal   MedicalStatus = "NORMAL"
	MedicalStatusConcern  MedicalStatus = "CONCERN"
	MedicalStatusUrgent   MedicalStatus = "URGENT"
	MedicalStatusCritical MedicalStatus = "CRITICAL"
)
