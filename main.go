// gcloudctx is a fast command-line tool for switching between gcloud configurations.
// Inspired by kubectx, it provides an intuitive interface for managing multiple
// Google Cloud SDK configurations with features like interactive selection,
// previous configuration switching, and ADC synchronization.
package main

import "github.com/Okabe-Junya/gcloudctx/cmd"

func main() {
	cmd.Execute()
}
