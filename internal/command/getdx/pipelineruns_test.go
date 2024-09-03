package getdx_test

import (
	"github.com/cirruslabs/cirrus-webhooks-server/internal/command/datadog/payload"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/command/getdx"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPipelineRunsRequestEnrichment(t *testing.T) {
	var payload payload.BuildOrTask

	buildID := int64(42)
	payload.Build.ID = &buildID

	taskName := "test task"
	payload.Task.Name = &taskName

	taskLocalGroupID := int64(7)
	payload.Task.LocalGroupID = &taskLocalGroupID

	taskStatusTimestamp := int64(3)
	payload.Task.StatusTimestamp = &taskStatusTimestamp

	actualPipelineRunsRequest := getdx.PipelineRunsRequest{
		PipelineSource: "Cirrus CI",
	}

	require.NoError(t, actualPipelineRunsRequest.Enrich(&payload))
	require.Equal(t, getdx.PipelineRunsRequest{
		PipelineName:   "test task",
		PipelineSource: "Cirrus CI",
		ReferenceID:    "build-42-local-group-id-7",
		StartedAt:      "3",
	}, actualPipelineRunsRequest)
}
