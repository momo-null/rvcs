package repository

import "gorm.io/gorm"

type Repositories struct {
	User             UserRepository
	Device           DeviceRepository
	DeviceSession    DeviceSessionRepository
	ActivityLog      ActivityLogRepository
	Stream           StreamRepository
	Alert            AlertRepository
	AlertRule        AlertRuleRepository
	DeviceCommand    DeviceCommandRepository
	RegistrationCode RegistrationCodeRepository
	DB               *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:             NewUserRepository(db),
		Device:           NewDeviceRepository(db),
		DeviceSession:    NewDeviceSessionRepository(db),
		ActivityLog:      NewActivityLogRepository(db),
		Stream:           NewStreamRepository(db),
		Alert:            NewAlertRepository(db),
		AlertRule:        NewAlertRuleRepository(db),
		DeviceCommand:    NewDeviceCommandRepository(db),
		RegistrationCode: NewRegistrationCodeRepository(db),
		DB:               db,
	}
}
