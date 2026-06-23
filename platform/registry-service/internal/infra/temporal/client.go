package temporal

import (
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
)

// NewClient builds a Temporal client from the TEMPORAL_HOST environment variable.
// It defaults to localhost:7233 when the variable is not set.
func NewClient() (client.Client, error) {
	host := os.Getenv("TEMPORAL_HOST")
	if host == "" {
		host = "localhost:7233"
	}
	c, err := client.Dial(client.Options{HostPort: host})
	if err != nil {
		return nil, fmt.Errorf("dial temporal %q: %w", host, err)
	}
	return c, nil
}
