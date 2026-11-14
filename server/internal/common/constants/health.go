package constants

// Overall service status reported by the health endpoint.
const (
	HealthStatusOK       = "ok"
	HealthStatusDegraded = "degraded"
)

// Database status reported by the health endpoint.
const (
	DBStatusUp   = "up"
	DBStatusDown = "down"
)
