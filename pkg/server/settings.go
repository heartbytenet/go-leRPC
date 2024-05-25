package server

type Settings struct {
	Port uint16
}

func NewSettingsDefault() Settings {
	return Settings{
		Port: 3000,
	}
}
