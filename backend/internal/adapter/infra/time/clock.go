package time

import (
	"time"

	"chronome/internal/usecase/provider"
)

// SystemClock は OS の時計で provider.Clock を実装する。
type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

var _ provider.Clock = SystemClock{}
