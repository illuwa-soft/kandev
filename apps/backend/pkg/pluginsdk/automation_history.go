package pluginsdk

import "context"

type AutomationHistoryQuery struct {
	RequestID, WorkspaceID, AutomationID string
	Page                                 ExactReadPage
}

type AutomationRunLookup struct{ RequestID, WorkspaceID, RunID string }

type AutomationSchedule struct {
	TriggerID   string `json:"trigger_id"`
	TriggerType string `json:"trigger_type"`
	Expression  string `json:"expression,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	Enabled     bool   `json:"enabled"`
}

type AutomationDefinition struct {
	ID                string               `json:"id"`
	WorkspaceID       string               `json:"workspace_id"`
	Name              string               `json:"name"`
	Enabled           bool                 `json:"enabled"`
	AgentProfileID    string               `json:"agent_profile_id,omitempty"`
	ExecutorProfileID string               `json:"executor_profile_id,omitempty"`
	CreatedAt         string               `json:"created_at"`
	UpdatedAt         string               `json:"updated_at"`
	Schedules         []AutomationSchedule `json:"schedules"`
	ResourceVersion   string               `json:"resource_version"`
	ObservedAt        string               `json:"observed_at"`
}

type AutomationRunObservation struct {
	ID               string `json:"id"`
	WorkspaceID      string `json:"workspace_id"`
	AutomationID     string `json:"automation_id"`
	TriggerID        string `json:"trigger_id,omitempty"`
	TriggerType      string `json:"trigger_type"`
	TaskID           string `json:"task_id,omitempty"`
	SessionID        string `json:"session_id,omitempty"`
	TurnID           string `json:"turn_id,omitempty"`
	Status           string `json:"status"`
	Summary          string `json:"summary,omitempty"`
	ErrorMessage     string `json:"error_message,omitempty"`
	RetryGroupID     string `json:"retry_group_id,omitempty"`
	RetryParentRunID string `json:"retry_parent_run_id,omitempty"`
	AttemptNumber    int64  `json:"attempt_number"`
	CreatedAt        string `json:"created_at"`
	ResourceVersion  string `json:"resource_version"`
	ObservedAt       string `json:"observed_at"`
}

type AutomationDefinitionPage struct {
	Items    []AutomationDefinition
	PageInfo ExactReadPageInfo
}

type AutomationRunPage struct {
	Items    []AutomationRunObservation
	PageInfo ExactReadPageInfo
}

type AutomationHistoryReader interface {
	ListAutomations(context.Context, AutomationHistoryQuery) (AutomationDefinitionPage, error)
	ListRuns(context.Context, AutomationHistoryQuery) (AutomationRunPage, error)
	GetRun(context.Context, AutomationRunLookup) (AutomationRunObservation, HostReadReceipt, error)
}

type AutomationHistoryHost interface {
	AutomationHistory() AutomationHistoryReader
}

// AutomationHistory returns an optional reader; its RPCs can report Unimplemented on older Hosts.
func AutomationHistory(host Host) (AutomationHistoryReader, bool) {
	provider, ok := host.(AutomationHistoryHost)
	if !ok || provider.AutomationHistory() == nil {
		return nil, false
	}
	return provider.AutomationHistory(), true
}
