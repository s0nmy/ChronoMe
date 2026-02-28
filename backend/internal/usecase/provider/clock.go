package provider

import "time"

// Clock はユースケースで使う時刻取得元を抽象化する。
type Clock interface {
	Now() time.Time
}
