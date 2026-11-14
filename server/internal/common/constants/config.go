package constants

import "time"

const (
	HealthPingTimeout  = 2 * time.Second
	StartupPingTimeout = 5 * time.Second
	ShutdownTimeout    = 10 * time.Second
)
