package datadog

import (
	"encoding/json"
	"errors"
	"fmt"
	payloadpkg "github.com/cirruslabs/cirrus-webhooks-server/internal/command/datadog/payload"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/datadogsender"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"time"
)

var dogstatsdAddr string
var apiKey string
var apiSite string

var (
	ErrDatadogFailed = errors.New("failed to stream Cirrus CI events to Datadog")
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datadog",
		Short: "Stream Cirrus CI webhook events to Datadog",
		RunE:  run,
	}

	server.AppendFlags(cmd)

	cmd.PersistentFlags().StringVar(&dogstatsdAddr, "dogstatsd-addr", "",
		"enables sending webhook events as Datadog events via the DogStatsD protocol to the specified address "+
			"(for example, --dogstatsd-addr=127.0.0.1:8125)")
	cmd.PersistentFlags().StringVar(&apiKey, "api-key", "",
		"enables sending webhook events as Datadog logs via the Datadog API using the specified API key")
	cmd.PersistentFlags().StringVar(&apiSite, "api-site", "datadoghq.com",
		"specifies the Datadog site to use when sending webhook events as Datadog logs via the Datadog API")

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	// Initialize a Datadog sender
	var sender datadogsender.Sender
	var err error

	switch {
	case dogstatsdAddr != "":
		sender, err = datadogsender.NewDogstatsdSender(dogstatsdAddr)
	case apiKey != "":
		sender, err = datadogsender.NewAPISender(apiKey, apiSite)
	default:
		return fmt.Errorf("%w: no sender configured, please specify either --api-key or --dogstatsd-addr",
			ErrDatadogFailed)
	}

	if err != nil {
		return err
	}

	return server.New(func(ctx echo.Context, presentedEventType string, body []byte, logger *zap.SugaredLogger) error {
		return processWebhookEvent(ctx, presentedEventType, body, sender, logger)
	}, zap.S()).Run(cmd.Context())
}

func processWebhookEvent(
	ctx echo.Context,
	presentedEventType string,
	body []byte,
	sender datadogsender.Sender,
	logger *zap.SugaredLogger,
) error {
	// Decode the event
	var payload payloadpkg.Payload

	switch presentedEventType {
	case "audit_event":
		payload = &payloadpkg.AuditEvent{}
	case "build", "task":
		payload = &payloadpkg.BuildOrTask{}
	default:
		return nil
	}

	if err := json.Unmarshal(body, payload); err != nil {
		return fmt.Errorf("failed to enrich Datadog event with tags: "+
			"failed to parse the webhook event of type %q as JSON: %v", presentedEventType, err)
	}

	// Create a new Datadog event and enrich it with tags
	evt := &datadogsender.Event{
		Title: "Webhook event",
		Text:  string(body),
		Tags:  []string{fmt.Sprintf("webhook_event_type:%s", presentedEventType)},
	}

	payload.Enrich(ctx.Request().Header, evt, logger)

	// Datadog silently discards log events submitted with a
	// timestamp that is more than 18 hours in the past, sigh.
	//
	// [1]: https://docs.datadoghq.com/api/latest/logs/#send-logs
	if !evt.Timestamp.IsZero() && time.Since(evt.Timestamp) >= 18*time.Hour {
		logger.Warnf("submitting an event of type %q with a timestamp that is more than "+
			"18 hours in the past, it'll likely going to be discarded", presentedEventType)
	}

	// Log this event to Datadog
	if err := sender.SendEvent(ctx.Request().Context(), evt); err != nil {
		return fmt.Errorf("%w: %v", ErrDatadogFailed, err)
	}

	return nil
}
