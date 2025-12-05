package enums

type UserRole string

const (
	UserRoleAdmin   UserRole = "ADMIN"
	UserRoleDoctor  UserRole = "DOCTOR"
	UserRolePatient UserRole = "PATIENT"
)

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)
