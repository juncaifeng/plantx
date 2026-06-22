package config

// Loader loads configuration values from a source.
type Loader interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetStringSlice(key string) []string
}
