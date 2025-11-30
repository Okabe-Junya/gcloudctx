// Package gcloud provides functionality to interact with Google Cloud SDK configurations.
// It wraps the gcloud CLI commands and provides a convenient Go interface for managing
// configurations, activating them, and synchronizing Application Default Credentials.
package gcloud

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

// CoreProperties represents core configuration properties
type CoreProperties struct {
	Account            string `json:"account,omitempty"`
	Project            string `json:"project,omitempty"`
	DisableUsageReport bool   `json:"disable_usage_reporting,omitempty"`
}

// ComputeProperties represents compute configuration properties
type ComputeProperties struct {
	Region string `json:"region,omitempty"`
	Zone   string `json:"zone,omitempty"`
}
