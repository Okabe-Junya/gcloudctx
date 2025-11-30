package interactive

// Environment variable names for customizing fzf behavior
const (
	// EnvIgnoreFzf skips fzf and shows current config when no args are provided
	EnvIgnoreFzf = "GCLOUDCTX_IGNORE_FZF"

	// EnvFzfHeight controls the height of the fzf window
	EnvFzfHeight = "GCLOUDCTX_FZF_HEIGHT"

	// EnvFzfPreviewWindow controls the preview window position and size
	EnvFzfPreviewWindow = "GCLOUDCTX_FZF_PREVIEW_WINDOW"

	// EnvDisablePreview disables preview in interactive mode when set to "1"
	EnvDisablePreview = "GCLOUDCTX_DISABLE_PREVIEW"

	// EnvFzfOptions allows additional fzf options to be specified
	EnvFzfOptions = "GCLOUDCTX_FZF_OPTIONS"
)

// Default values for fzf options
const (
	DefaultFzfHeight        = "40%"
	DefaultFzfPreviewWindow = "right:50%:wrap"
)

// Command names
const (
	// PreviewCommand is the internal command used for fzf preview
	PreviewCommand = "__preview"
)
