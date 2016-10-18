package model

type PassivationConfig struct {
	DelayInSeconds int    `json:"delayInSeconds,omitempty"`
	Enabled        bool   `json:"enabled,omitempty"`
	Action         string `json:"action,omitempty"`
}

func DefaultPassivation() *PassivationConfig {
	return &PassivationConfig{
		DelayInSeconds: 3600 * 12,
		Enabled:        true,
		Action:         "passivate",
	}
}
