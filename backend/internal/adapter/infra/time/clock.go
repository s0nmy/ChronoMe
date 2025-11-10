package time

import (
	"time"

	"chronome/internal/usecase/provider"
)

// SystemClock implements provider.Clock using the OS clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

var _ provider.Clock = SystemClock{}
