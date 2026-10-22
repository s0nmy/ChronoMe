package repository

import "errors"

// ErrConflict は取得後に対象データが更新または削除されたことを示す。
var ErrConflict = errors.New("resource changed; reload before retrying")

// ErrNotFound は認証ユーザーの操作対象が存在しないことを示す。
var ErrNotFound = errors.New("resource not found")
