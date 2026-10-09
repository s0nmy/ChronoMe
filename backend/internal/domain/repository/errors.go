package repository

import "errors"

// ErrConflict means the resource changed or was deleted after it was read.
var ErrConflict = errors.New("resource changed; reload before retrying")
