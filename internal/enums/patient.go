package enums

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelMedium   RiskLevel = "MEDIUM"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
)

type MonitoringFrequency string

const (
	MonitoringFrequencyTwiceDaily    MonitoringFrequency = "TWICE_DAILY"
	MonitoringFrequencyDaily         MonitoringFrequency = "DAILY"
	MonitoringFrequencyEveryOtherDay MonitoringFrequency = "EVERY_OTHER_DAY"
	MonitoringFrequencyWeekly        MonitoringFrequency = "WEEKLY"
)

type PatientStatus string

const (
	PatientStatusActive     PatientStatus = "ACTIVE"
	PatientStatusPaused     PatientStatus = "PAUSED"
	PatientStatusDischarged PatientStatus = "DISCHARGED"
	PatientStatusCritical   PatientStatus = "CRITICAL"
)
