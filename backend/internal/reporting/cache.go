package reporting

import "time"

// Cache is the reporting-specific cache interface (Cache-Aside pattern).
// A nil implementation means no caching.
type Cache interface {
	Get(key string) ([]byte, bool, error)
	Set(key string, value []byte, ttl time.Duration) error
	Delete(key string) error
}
