package provider

import "time"

// Clock abstracts the source of time for usecases.
type Clock interface {
	Now() time.Time
}
