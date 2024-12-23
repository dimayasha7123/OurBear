package config

import "time"

type Config struct {
	ApiKey  string
	Delay   time.Duration
	Timeout time.Duration
}
