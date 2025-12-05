package enums

type ScheduleFrequency string

const (
	ScheduleFrequencyTwiceDaily    ScheduleFrequency = "TWICE_DAILY"
	ScheduleFrequencyDaily         ScheduleFrequency = "DAILY"
	ScheduleFrequencyEveryOtherDay ScheduleFrequency = "EVERY_OTHER_DAY"
	ScheduleFrequencyWeekly        ScheduleFrequency = "WEEKLY"
)
