package enums

type VitalType string

const (
	VitalTypeBloodPressure    VitalType = "BLOOD_PRESSURE"
	VitalTypeGlucose          VitalType = "GLUCOSE"
	VitalTypeHeartRate        VitalType = "HEART_RATE"
	VitalTypeTemperature      VitalType = "TEMPERATURE"
	VitalTypeWeight           VitalType = "WEIGHT"
	VitalTypeOxygenSaturation VitalType = "OXYGEN_SATURATION"
)

type VitalUnit string

const (
	VitalUnitMmHg    VitalUnit = "MMHG"
	VitalUnitMgDL    VitalUnit = "MG/DL"
	VitalUnitBPM     VitalUnit = "BPM"
	VitalUnitCelsius VitalUnit = "°C"
	VitalUnitKg      VitalUnit = "KG"
	VitalUnitPercent VitalUnit = "%"
)
