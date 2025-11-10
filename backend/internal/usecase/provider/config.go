package provider

import "time"

// AppConfig exposes configuration needed by usecases.
type AppConfig interface {
	DefaultProjectColor() string
	SessionTTL() time.Duration
}
