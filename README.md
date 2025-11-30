# gcloudctx

Fast way to switch between gcloud configurations, inspired by [kubectx](https://github.com/ahmetb/kubectx)

## Features

- **Fast switching** between gcloud configurations
- **Interactive mode** with [fzf](https://github.com/junegunn/fzf) integration and preview
- **Previous configuration** switching with `-` (dash)
- **ADC synchronization** support
- **Configuration management** (create, delete, rename)
- **Colorful output** for better readability
- **Customizable fzf** options via environment variables
- **Cross-platform** support (Linux, macOS, Windows)

## Installation

### Go Install

```bash
go install github.com/Okabe-Junya/gcloudctx@latest
```

## Usage

### Basic Commands

```bash
# Interactive mode with fzf (if fzf is installed, launches automatically)
gcloudctx

# List all configurations
gcloudctx -l
gcloudctx --list

# Switch to a specific configuration
gcloudctx my-config

# Switch to previous configuration
gcloudctx -

# Force interactive mode with fzf
gcloudctx -i
gcloudctx --interactive

# Show current configuration (skip fzf even if installed)
gcloudctx -c
gcloudctx --current

# Show detailed configuration information
gcloudctx --info

# Disable colored output
gcloudctx --no-color

# Create a new configuration
gcloudctx create my-new-config
gcloudctx create my-new-config --activate

# Delete a configuration
gcloudctx delete my-old-config
gcloudctx delete my-old-config --force

# Rename a configuration
gcloudctx rename old-name new-name
```

### Advanced Features

#### ADC Synchronization

Sync Application Default Credentials when switching configurations:

```bash
# Switch and sync ADC
gcloudctx my-config --sync-adc

# Sync ADC with service account impersonation
gcloudctx my-config --sync-adc --impersonate-service-account=sa@project.iam.gserviceaccount.com
```

**⚠️ Security Warning:**
- ADC synchronization will trigger an OAuth flow and store credentials in `~/.config/gcloud/application_default_credentials.json`
- These credentials have broad access to GCP resources. Never commit this file to version control
- Use `--impersonate-service-account` in production to limit credential scope
- Review the [principle of least privilege](https://cloud.google.com/iam/docs/using-iam-securely#least_privilege) when granting permissions

#### Configuration Management

Create, delete, and rename configurations:

```bash
# Create a new configuration
gcloudctx create production --activate

# Delete an unused configuration
gcloudctx delete old-project

# Rename a configuration
gcloudctx rename dev development
```

## Important Notes on ADC

**Application Default Credentials (ADC) are independent from gcloud configurations.**

- Switching configurations does NOT automatically switch ADC
- Use `--sync-adc` flag to explicitly sync ADC when needed
- ADC is stored in `~/.config/gcloud/application_default_credentials.json`

This design follows GCP's architecture where:

- gcloud configurations: for CLI operations
- ADC: for application/SDK authentication

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
