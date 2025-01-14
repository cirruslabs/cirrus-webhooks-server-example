package datadogsender

import (
	"context"
	"errors"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"strings"
	"time"
)

var ErrAPISenderFailed = errors.New("API sender failed to send the event")

type APISender struct {
	apiClient *datadog.APIClient
	logsAPI   *datadogV2.LogsApi

	apiKey  string
	apiSite string
}

func NewAPISender(apiKey string, apiSite string) (*APISender, error) {
	apiClient := datadog.NewAPIClient(datadog.NewConfiguration())

	return &APISender{
		apiClient: apiClient,
		logsAPI:   datadogV2.NewLogsApi(apiClient),

		apiKey:  apiKey,
		apiSite: apiSite,
	}, nil
}

func (sender *APISender) SendEvent(ctx context.Context, event *Event) error {
	ctx = context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: sender.apiKey,
			},
		},
	)

	ctx = context.WithValue(ctx,
		datadog.ContextServerVariables,
		map[string]string{
			"site": sender.apiSite,
		})

	logItem := datadogV2.HTTPLogItem{
		Ddsource: datadog.PtrString("Cirrus Webhooks Server"),
		Ddtags:   datadog.PtrString(strings.Join(event.Tags, ",")),
		Message:  event.Text,
	}

	if !event.Timestamp.IsZero() {
		logItem.AdditionalProperties = map[string]string{
			// https://docs.datadoghq.com/service_management/events/pipelines_and_processors/date_remapper/
			"timestamp": event.Timestamp.Format(time.RFC3339),
		}
	}

	_, _, err := sender.logsAPI.SubmitLog(ctx, []datadogV2.HTTPLogItem{logItem})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAPISenderFailed, err)
	}

	return nil
}
