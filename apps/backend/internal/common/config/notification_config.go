package config

import "github.com/kandev/kandev/internal/common/weburl"

type NotificationsConfig struct {
	PublicBaseURL string `mapstructure:"publicBaseUrl"`
}

func (c NotificationsConfig) validate() error {
	_, err := weburl.ParseOrigin(c.PublicBaseURL)
	return err
}
