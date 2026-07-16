package provider

// AppConfig はユースケースが必要とする設定を公開する。
type AppConfig interface {
	DefaultProjectColor() string
}
