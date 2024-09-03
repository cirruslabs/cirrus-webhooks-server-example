package getdx

import (
	"fmt"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/command/datadog/payload"
	"strconv"
)

type PipelineRunsStatus string

const (
	PipelineRunsStatusFailure   PipelineRunsStatus = "failure"
	PipelineRunsStatusRunning   PipelineRunsStatus = "running"
	PipelineRunsStatusSuccess   PipelineRunsStatus = "success"
	PipelineRunsStatusCancelled PipelineRunsStatus = "cancelled"
)

type PipelineRunsRequest struct {
	// PipelineName describes the name of the pipeline that this job is for.
	PipelineName string `json:"pipeline_name"`

	// PipelineSource describes which project source is this? CircleCI, Jenkins, Tekton etc.
	PipelineSource string `json:"pipeline_source"`

	// ReferenceID is a globally unique identifier for this run.
	ReferenceID string `json:"reference_id"`

	// StartedAt is string with a Unix timestamp. Each notification does not need to have
	// the same StartedAt value. They can arrive in any sequence and the earliest value
	// will be used.
	StartedAt string `json:"started_at"`

	// Status is either a "failure", "running", "success" or "cancelled".
	//
	// When not specified, defaults to "unknown".
	Status PipelineRunsStatus `json:"status"`

	// FinishedAt is a string with Unix timestamp.
	FinishedAt string `json:"finished_at"`

	// Repository is the full name of repository.
	Repository string `json:"repository"`

	// CommitSHA that the build was run on. If CommitSHA is provided, Repository is required.
	CommitSHA string `json:"commit_sha"`

	// CommitSHA is the number of the pull request that the build was run on.
	// If PRNumber is provided, Repository is required.
	PRNumber int64 `json:"pr_number"`

	// SourceURL is a web address to view details about the pipeline run.
	SourceURL string `json:"source_url"`

	// HeadBranch is used to know which branch the pipeline run was run against.
	HeadBranch string `json:"head_branch"`

	// Email is an email address of the user who initiated the build.
	Email string `json:"email"`

	// GithubUsername is a GitHub username of the user who initiated the build.
	GithubUsername string `json:"github_username"`
}

func (pipelineRunsRequest *PipelineRunsRequest) Enrich(payload *payload.BuildOrTask) error {
	if value := payload.Task.Name; value != nil {
		pipelineRunsRequest.PipelineName = *value
	} else {
		return fmt.Errorf("\"pipeline_name\" field is required, but no task name found in the webhook payload")
	}

	if payload.Build.ID != nil && payload.Task.LocalGroupID != nil {
		pipelineRunsRequest.ReferenceID = fmt.Sprintf("build-%d-local-group-id-%d",
			*payload.Build.ID, *payload.Task.LocalGroupID)
	} else {
		return fmt.Errorf("\"reference_id\" field is required, but no build ID and/or task's " +
			"local group ID found in the webhook payload")
	}

	if value := payload.Task.StatusTimestamp; value != nil {
		pipelineRunsRequest.StartedAt = strconv.FormatInt(*value, 10)
	} else {
		return fmt.Errorf("\"started_at\" field is required, but no task status timestamp found in the webhook payload")
	}

	if value := payload.Task.Status; value != nil {
		switch *value {
		case "EXECUTING":
			pipelineRunsRequest.Status = PipelineRunsStatusRunning
		case "FAILED", "ERRORED":
			pipelineRunsRequest.Status = PipelineRunsStatusFailure
		case "COMPLETED":
			pipelineRunsRequest.Status = PipelineRunsStatusSuccess
		case "ABORTED":
			pipelineRunsRequest.Status = PipelineRunsStatusCancelled
		}
	}

	if payload.Repository.Owner != nil && payload.Repository.Name != nil {
		pipelineRunsRequest.Repository = fmt.Sprintf("%s/%s",
			*payload.Repository.Owner, *payload.Repository.Name)
	}

	if value := payload.Build.ChangeIDInRepo; value != nil {
		pipelineRunsRequest.CommitSHA = *value
	}

	if value := payload.Build.PullRequest; value != nil {
		pipelineRunsRequest.PRNumber = *value
	}

	if value := payload.Task.ID; value != nil {
		pipelineRunsRequest.SourceURL = fmt.Sprintf("https://cirrus-ci.com/task/%d", *value)
	}

	if value := payload.Build.User.Username; value != nil {
		pipelineRunsRequest.GithubUsername = *value
	}

	return nil
}
