package external

import (
	"context"
	"errors"
)

// SettingReader is the minimal interface the integration clients need to look
// up their current configuration at request time. *appsetting.Service satisfies
// it implicitly. Defining the interface here (where it is consumed) keeps
// external independent of the application layer.
type SettingReader interface {
	Get(ctx context.Context, key string) (string, error)
}

// ErrNotConfigured is returned by integration clients when the required
// setting keys are empty in storage. Handlers translate it into a
// "not configured" response so configuration changes made through the UI
// take effect on the next request without restarting the API.
var ErrNotConfigured = errors.New("integration not configured")

// IsNotConfigured reports whether err indicates the integration has no URL
// or API key configured in settings.
func IsNotConfigured(err error) bool { return errors.Is(err, ErrNotConfigured) }
