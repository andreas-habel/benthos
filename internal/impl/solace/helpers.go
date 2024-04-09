package solace

import (
	"strings"

	"solace.dev/go/messaging/pkg/solace/config"
)

// stripPropertyName returns the last part from config property names to be used
// in benthos configurations
func stripPropertyName(propertyName config.ServiceProperty) (shortConfig string) {
	prop := string(propertyName)
	shortConfig = prop[strings.LastIndex(prop, ".")+1:]
	return strings.ToLower(strings.ReplaceAll(shortConfig, "-", "_"))
}
