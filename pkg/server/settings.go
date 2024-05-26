package server

import "time"

type Settings struct {
	Port          uint16
	ExecutorLimit int
	ExecutorDelay time.Duration
}

func NewSettingsDefault() Settings {
	return Settings{
		Port:          3000,
		ExecutorLimit: 65536,
		ExecutorDelay: 1,
	}
}
