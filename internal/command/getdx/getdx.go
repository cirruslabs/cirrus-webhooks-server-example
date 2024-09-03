package getdx

import (
	"bytes"
	"encoding/json"
	"fmt"
	payloadpkg "github.com/cirruslabs/cirrus-webhooks-server/internal/command/datadog/payload"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"net/http"
)

var dxInstance string
var dxAPIKey string

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getdx",
		Short: "Stream Cirrus CI webhook events to DX's Data Cloud API",
		RunE:  run,
	}

	server.AppendFlags(cmd, "task")

	cmd.PersistentFlags().StringVar(&dxInstance, "dx-instance", "",
		"DX instance to use when sending webhook events as DX Pipeline events to the Data Cloud API")
	cmd.PersistentFlags().StringVar(&dxAPIKey, "dx-api-key", "",
		"API key to use when sending webhook events as DX Pipeline events to the Data Cloud API")

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	if dxInstance == "" {
		return fmt.Errorf("\"--dx-instance\" is required")
	}

	return server.New(processWebhookEvent, zap.S()).Run(cmd.Context())
}

func processWebhookEvent(ctx echo.Context, presentedEventType string, body []byte, logger *zap.SugaredLogger) error {
	// Decode the event
	var payload payloadpkg.BuildOrTask

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to parse the webhook event of type %q as JSON: %w",
			presentedEventType, err)
	}

	pipelineRunsRequest := PipelineRunsRequest{
		PipelineSource: "Cirrus CI",
	}

	if err := pipelineRunsRequest.Enrich(&payload); err != nil {
		return fmt.Errorf("failed to enrich GetDX event: %w", err)
	}

	pipelineRunsReqeuestJSON, err := json.Marshal(&pipelineRunsRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal GetDX event as JSON: %w", err)
	}

	url := fmt.Sprintf("https://%s.getdx.net/api/pipelineRuns.sync", dxInstance)

	req, err := http.NewRequestWithContext(ctx.Request().Context(), http.MethodPost, url,
		bytes.NewReader(pipelineRunsReqeuestJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DX's Data Cloud API unexpectedly responded with HTTP %d",
			resp.StatusCode)
	}

	return nil
}
