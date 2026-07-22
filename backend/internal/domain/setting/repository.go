package setting

import "context"

// Repository is the persistence port for application settings. Settings are
// a flat key/value store: keys are namespaced strings such as "ui.theme" or
// "integration.openai_api_key".
type Repository interface {
	// Get returns the stored value for key, or an empty string when the
	// key is not set. A missing key is not an error.
	Get(ctx context.Context, key string) (string, error)

	// GetAll returns every stored setting keyed by name.
	GetAll(ctx context.Context) (map[string]string, error)

	// Set writes value for key, overwriting any existing value.
	Set(ctx context.Context, key, value string) error

	// Seed writes value for key only when the key is not already set, so
	// environment defaults never overwrite user-configured values.
	Seed(ctx context.Context, key, value string) error
}
