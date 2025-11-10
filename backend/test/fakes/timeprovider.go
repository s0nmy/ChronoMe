package fakes

import (
	"time"

	"chronome/internal/usecase/provider"
)

// FixedTimeProvider allows tests to control Now().
type FixedTimeProvider struct {
	NowFunc func() time.Time
}

func (f FixedTimeProvider) Now() time.Time {
	if f.NowFunc != nil {
		return f.NowFunc().UTC()
	}
	return time.Unix(0, 0).UTC()
}

var _ provider.Clock = FixedTimeProvider{}
