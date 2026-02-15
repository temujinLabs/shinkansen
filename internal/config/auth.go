package config

import (
	"encoding/base64"
	"fmt"
)

// BasicAuthHeader returns the Authorization header value for Jira API token auth.
func BasicAuthHeader(email, token string) string {
	credentials := fmt.Sprintf("%s:%s", email, token)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}
