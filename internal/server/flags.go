package server

import "github.com/spf13/cobra"

var httpAddr string
var httpPath string
var eventTypes []string
var secretToken string

func AppendFlags(cmd *cobra.Command, specificEventTypes ...string) {
	cmd.Flags().StringVar(&httpAddr, "http-addr", ":8080",
		"address on which the HTTP server will listen on")
	cmd.Flags().StringVar(&httpPath, "http-path", "/",
		"HTTP path on which the webhook events will be expected")
	cmd.Flags().StringVar(&secretToken, "secret-token", "",
		"if specified, this value will be used as a HMAC SHA-256 secret to verify the webhook events")

	if len(specificEventTypes) != 0 {
		eventTypes = specificEventTypes
	} else {
		cmd.Flags().StringSliceVar(&eventTypes, "event-types", []string{},
			"comma-separated list of the event types to limit processing to "+
				"(for example, --event-types=audit_event or --event-types=build,task")
	}
}
