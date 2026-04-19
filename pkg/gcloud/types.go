// Package gcloud provides functionality to interact with Google Cloud SDK configurations.
// It wraps the gcloud CLI commands and provides a convenient Go interface for managing
// configurations, activating them, and synchronizing Application Default Credentials.
package gcloud

import (
	"encoding/json"
	"strings"
)

// Configuration represents a gcloud configuration
type Configuration struct {
	Name       string     `json:"name"`
	IsActive   bool       `json:"is_active"`
	Properties Properties `json:"properties"`
}

// Properties represents configuration properties
type Properties struct {
	Core    CoreProperties    `json:"core,omitempty"`
	Compute ComputeProperties `json:"compute,omitempty"`
}

// Bool handles gcloud's JSON output where boolean values may be
// represented as strings ("True"/"False") instead of native JSON booleans.
type Bool bool

// UnmarshalJSON implements json.Unmarshaler, accepting both native JSON
// booleans and the string representations ("True"/"False") that gcloud emits.
func (b *Bool) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch v := raw.(type) {
	case bool:
		*b = Bool(v)
	case string:
		*b = Bool(strings.EqualFold(v, "true"))
	default:
		*b = false
	}
	return nil
}

// CoreProperties represents core configuration properties
type CoreProperties struct {
	Account            string `json:"account,omitempty"`
	Project            string `json:"project,omitempty"`
	DisableUsageReport Bool   `json:"disable_usage_reporting,omitempty"`
}

// ComputeProperties represents compute configuration properties
type ComputeProperties struct {
	Region string `json:"region,omitempty"`
	Zone   string `json:"zone,omitempty"`
}
