package interactive

import (
	"fmt"
	"strings"
)

// ParseConfigurationName extracts the configuration name from a formatted line
// Expected formats:
//   - "* config-name (account) [project]" (active)
//   - "  config-name (account) [project]" (non-active)
func ParseConfigurationName(line string) (string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", fmt.Errorf("empty line")
	}

	// Split the line into fields
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid format: no fields found")
	}

	// Find the configuration name (skip marker and parenthesized/bracketed fields)
	for _, part := range parts {
		// Skip marker and fields that start with ( or [
		if part == "*" || strings.HasPrefix(part, "(") || strings.HasPrefix(part, "[") {
			continue
		}
		// This should be the configuration name
		return part, nil
	}

	return "", fmt.Errorf("could not extract configuration name")
}
