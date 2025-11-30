package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for gcloudctx.

To load completions:

Bash:
  $ source <(gcloudctx completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ gcloudctx completion bash > /etc/bash_completion.d/gcloudctx
  # macOS:
  $ gcloudctx completion bash > /usr/local/etc/bash_completion.d/gcloudctx

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ gcloudctx completion zsh > "${fpath[1]}/_gcloudctx"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ gcloudctx completion fish | source
  # To load completions for each session, execute once:
  $ gcloudctx completion fish > ~/.config/fish/completions/gcloudctx.fish

PowerShell:
  PS> gcloudctx completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> gcloudctx completion powershell > gcloudctx.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
