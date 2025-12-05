package enums

type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "LOW"
	AlertSeverityMedium   AlertSeverity = "MEDIUM"
	AlertSeverityHigh     AlertSeverity = "HIGH"
	AlertSeverityCritical AlertSeverity = "CRITICAL"
)

type AlertType string

const (
	AlertTypeVitalAbnormal     AlertType = "VITAL_ABNORMAL"
	AlertTypeNoResponse        AlertType = "NO_RESPONSE"
	AlertTypeSentimentNegative AlertType = "SENTIMENT_NEGATIVE"
	AlertTypePatternDetected   AlertType = "PATTERN_DETECTED"
)
