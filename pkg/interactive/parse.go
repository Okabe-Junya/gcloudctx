package interactive

import (
	"errors"
	"strings"
)

// ParseConfigurationName extracts the configuration name from a formatted line
// Expected formats:
//   - "* config-name (account) [project]" (active)
//   - "  config-name (account) [project]" (non-active)
func ParseConfigurationName(line string) (string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", errors.New("empty line")
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", errors.New("invalid format: no fields found")
	}

	for _, part := range parts {
		if part == "*" || strings.HasPrefix(part, "(") || strings.HasPrefix(part, "[") {
			continue
		}

		return part, nil
	}

	return "", errors.New("could not extract configuration name")
}
