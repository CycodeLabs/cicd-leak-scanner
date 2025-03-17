package github

import "fmt"

type WorkflowRunLogsExpiredError struct {
	WorkflowRunId int64
}

func (e *WorkflowRunLogsExpiredError) Error() string {
	return fmt.Sprintf("Workflow logs older than run %d have expired and are no longer available.", e.WorkflowRunId)
}
