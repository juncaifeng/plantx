// Package env provides a config.Loader that reads environment variables.
package env

import (
	"os"
	"strconv"
	"strings"

	"github.com/plantx/kit/kit-go/config"
)

// New creates an environment variable config loader.
// If prefix is non-empty, keys are looked up as PREFIX_KEY (upper-cased).
func New(prefix string) config.Loader {
	return &loader{prefix: strings.ToUpper(prefix)}
}

type loader struct {
	prefix string
}

func (l *loader) key(k string) string {
	if l.prefix == "" {
		return strings.ToUpper(k)
	}
	return l.prefix + "_" + strings.ToUpper(k)
}

func (l *loader) GetString(key string) string {
	return os.Getenv(l.key(key))
}

func (l *loader) GetInt(key string) int {
	v := l.GetString(key)
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

func (l *loader) GetBool(key string) bool {
	v := strings.ToLower(l.GetString(key))
	return v == "true" || v == "1" || v == "yes"
}

func (l *loader) GetStringSlice(key string) []string {
	v := l.GetString(key)
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}
