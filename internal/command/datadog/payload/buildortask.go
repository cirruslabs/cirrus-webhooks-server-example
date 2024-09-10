package payload

import (
	"fmt"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/datadogsender"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type BuildOrTask struct {
	Build struct {
		ID               *int64  `json:"id"`
		Status           *string `json:"status"`
		Branch           *string `json:"branch"`
		PullRequest      *int64  `json:"pullRequest"`
		PullRequestDraft *bool   `json:"pullRequestDraft"`
		ChangeIDInRepo   *string `json:"changeIdInRepo"`
		User             struct {
			Username *string `json:"username"`
		} `json:"user"`
	} `json:"build"`
	Task struct {
		ID               *int64   `json:"id"`
		Name             *string  `json:"name"`
		Status           *string  `json:"status"`
		StatusTimestamp  *int64   `json:"statusTimestamp"`
		InstanceType     *string  `json:"instanceType"`
		UniqueLabels     []string `json:"uniqueLabels"`
		ManualRerunCount *int64   `json:"manualRerunCount"`
		LocalGroupID     *int64   `json:"localGroupId"`
	} `json:"task"`

	common
}

func (buildOrTask BuildOrTask) Enrich(header http.Header, evt *datadogsender.Event, logger *zap.SugaredLogger) {
	buildOrTask.common.Enrich(header, evt, logger)

	if value := buildOrTask.Build.ID; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("build_id:%d", *value))
	}
	if value := buildOrTask.Build.Status; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("build_status:%s", *value))
	}
	if value := buildOrTask.Build.Branch; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("build_branch:%s", *value))
	}
	if value := buildOrTask.Build.PullRequest; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("build_pull_request:%d", *value))
	}
	if value := buildOrTask.Build.PullRequestDraft; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("build_pull_request_draft:%t", *value))
	}

	initializerUsername := "api"
	if value := buildOrTask.Build.User.Username; value != nil {
		initializerUsername = *value
	}
	evt.Tags = append(evt.Tags, fmt.Sprintf("initializer_username:%s", initializerUsername))

	if value := buildOrTask.Task.ID; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("task_id:%d", *value))
	}
	if value := buildOrTask.Task.Name; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("task_name:%s", *value))
	}
	if value := buildOrTask.Task.Status; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("task_status:%s", *value))
	}
	if value := buildOrTask.Task.InstanceType; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("task_instance_type:%s", *value))
	}
	if value := buildOrTask.Task.UniqueLabels; len(value) > 0 {
		evt.Tags = append(evt.Tags, fmt.Sprintf("task_unique_labels:%s", strings.Join(value, ",")))
	}
	if value := buildOrTask.Task.ManualRerunCount; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("manual_rerun_count:%d", *value))
	}
}
