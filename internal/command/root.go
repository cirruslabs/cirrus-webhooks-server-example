package command

import (
	"github.com/cirruslabs/cirrus-webhooks-server/internal/command/datadog"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/command/getdx"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/logginglevel"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

var debug bool

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "cws",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if debug {
				logginglevel.Level.SetLevel(zapcore.DebugLevel)
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	cmd.AddCommand(
		datadog.NewCommand(),
		getdx.NewCommand(),
	)

	return cmd
}
