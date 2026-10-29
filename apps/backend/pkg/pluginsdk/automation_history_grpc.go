package pluginsdk

import (
	"context"

	pluginv1 "github.com/kandev/kandev/proto/kandev/plugin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcAutomationHistoryReader struct{ client pluginv1.HostClient }

func (h *grpcHostClient) AutomationHistory() AutomationHistoryReader {
	return grpcAutomationHistoryReader{client: h.client}
}

func (s *grpcHostServer) automationHistoryReader() (AutomationHistoryReader, error) {
	reader, ok := AutomationHistory(s.impl)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "automation history reads are unavailable")
	}
	return reader, nil
}

func historyQueryToProto(q AutomationHistoryQuery) *pluginv1.AutomationHistoryQuery {
	return &pluginv1.AutomationHistoryQuery{RequestId: q.RequestID, WorkspaceId: q.WorkspaceID, AutomationId: q.AutomationID, Page: exactReadPageToProto(q.Page)}
}

func historyQueryFromProto(q *pluginv1.AutomationHistoryQuery) AutomationHistoryQuery {
	return AutomationHistoryQuery{RequestID: q.GetRequestId(), WorkspaceID: q.GetWorkspaceId(), AutomationID: q.GetAutomationId(), Page: exactReadPageFromProto(q.GetPage())}
}

func (r grpcAutomationHistoryReader) ListAutomations(ctx context.Context, q AutomationHistoryQuery) (AutomationDefinitionPage, error) {
	res, err := r.client.ListAutomationDefinitionsExact(ctx, historyQueryToProto(q))
	if err != nil {
		return AutomationDefinitionPage{}, err
	}
	items := make([]AutomationDefinition, len(res.GetItems()))
	for i, item := range res.GetItems() {
		items[i] = automationDefinitionFromProto(item)
	}
	return AutomationDefinitionPage{Items: items, PageInfo: exactReadPageInfoFromProto(res.GetPageInfo())}, nil
}

func (s *grpcHostServer) ListAutomationDefinitionsExact(ctx context.Context, q *pluginv1.AutomationHistoryQuery) (*pluginv1.AutomationDefinitionPage, error) {
	r, err := s.automationHistoryReader()
	if err != nil {
		return nil, err
	}
	page, err := r.ListAutomations(ctx, historyQueryFromProto(q))
	if err != nil {
		return nil, err
	}
	items := make([]*pluginv1.AutomationDefinition, len(page.Items))
	for i, item := range page.Items {
		items[i] = automationDefinitionToProto(item)
	}
	return &pluginv1.AutomationDefinitionPage{Items: items, PageInfo: exactReadPageInfoToProto(page.PageInfo)}, nil
}

func (r grpcAutomationHistoryReader) ListRuns(ctx context.Context, q AutomationHistoryQuery) (AutomationRunPage, error) {
	res, err := r.client.ListAutomationRunsExact(ctx, historyQueryToProto(q))
	if err != nil {
		return AutomationRunPage{}, err
	}
	items := make([]AutomationRunObservation, len(res.GetItems()))
	for i, item := range res.GetItems() {
		items[i] = automationRunFromProto(item)
	}
	return AutomationRunPage{Items: items, PageInfo: exactReadPageInfoFromProto(res.GetPageInfo())}, nil
}

func (s *grpcHostServer) ListAutomationRunsExact(ctx context.Context, q *pluginv1.AutomationHistoryQuery) (*pluginv1.AutomationRunPage, error) {
	r, err := s.automationHistoryReader()
	if err != nil {
		return nil, err
	}
	page, err := r.ListRuns(ctx, historyQueryFromProto(q))
	if err != nil {
		return nil, err
	}
	items := make([]*pluginv1.AutomationRunObservation, len(page.Items))
	for i, item := range page.Items {
		items[i] = automationRunToProto(item)
	}
	return &pluginv1.AutomationRunPage{Items: items, PageInfo: exactReadPageInfoToProto(page.PageInfo)}, nil
}

func (r grpcAutomationHistoryReader) GetRun(ctx context.Context, q AutomationRunLookup) (AutomationRunObservation, HostReadReceipt, error) {
	res, err := r.client.GetAutomationRunExact(ctx, &pluginv1.AutomationRunLookup{RequestId: q.RequestID, WorkspaceId: q.WorkspaceID, RunId: q.RunID})
	if err != nil {
		return AutomationRunObservation{}, HostReadReceipt{}, err
	}
	return automationRunFromProto(res.GetItem()), hostReadReceiptFromProto(res.GetReceipt()), nil
}

func (s *grpcHostServer) GetAutomationRunExact(ctx context.Context, q *pluginv1.AutomationRunLookup) (*pluginv1.AutomationRunResult, error) {
	r, err := s.automationHistoryReader()
	if err != nil {
		return nil, err
	}
	item, receipt, err := r.GetRun(ctx, AutomationRunLookup{RequestID: q.GetRequestId(), WorkspaceID: q.GetWorkspaceId(), RunID: q.GetRunId()})
	if err != nil {
		return nil, err
	}
	return &pluginv1.AutomationRunResult{Item: automationRunToProto(item), Receipt: hostReadReceiptToProto(receipt)}, nil
}

func automationDefinitionToProto(v AutomationDefinition) *pluginv1.AutomationDefinition {
	schedules := make([]*pluginv1.AutomationSchedule, len(v.Schedules))
	for i, s := range v.Schedules {
		schedules[i] = &pluginv1.AutomationSchedule{TriggerId: s.TriggerID, TriggerType: s.TriggerType, Expression: s.Expression, Timezone: s.Timezone, Enabled: s.Enabled}
	}
	return &pluginv1.AutomationDefinition{Id: v.ID, WorkspaceId: v.WorkspaceID, Name: v.Name, Enabled: v.Enabled, AgentProfileId: v.AgentProfileID, ExecutorProfileId: v.ExecutorProfileID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Schedules: schedules, ResourceVersion: v.ResourceVersion, ObservedAt: v.ObservedAt}
}

func automationDefinitionFromProto(v *pluginv1.AutomationDefinition) AutomationDefinition {
	schedules := make([]AutomationSchedule, len(v.GetSchedules()))
	for i, s := range v.GetSchedules() {
		schedules[i] = AutomationSchedule{TriggerID: s.GetTriggerId(), TriggerType: s.GetTriggerType(), Expression: s.GetExpression(), Timezone: s.GetTimezone(), Enabled: s.GetEnabled()}
	}
	return AutomationDefinition{ID: v.GetId(), WorkspaceID: v.GetWorkspaceId(), Name: v.GetName(), Enabled: v.GetEnabled(), AgentProfileID: v.GetAgentProfileId(), ExecutorProfileID: v.GetExecutorProfileId(), CreatedAt: v.GetCreatedAt(), UpdatedAt: v.GetUpdatedAt(), Schedules: schedules, ResourceVersion: v.GetResourceVersion(), ObservedAt: v.GetObservedAt()}
}

func automationRunToProto(v AutomationRunObservation) *pluginv1.AutomationRunObservation {
	return &pluginv1.AutomationRunObservation{Id: v.ID, WorkspaceId: v.WorkspaceID, AutomationId: v.AutomationID, TriggerId: v.TriggerID, TriggerType: v.TriggerType, TaskId: v.TaskID, SessionId: v.SessionID, TurnId: v.TurnID, Status: v.Status, Summary: v.Summary, ErrorMessage: v.ErrorMessage, RetryGroupId: v.RetryGroupID, RetryParentRunId: v.RetryParentRunID, AttemptNumber: v.AttemptNumber, CreatedAt: v.CreatedAt, ResourceVersion: v.ResourceVersion, ObservedAt: v.ObservedAt}
}

func automationRunFromProto(v *pluginv1.AutomationRunObservation) AutomationRunObservation {
	return AutomationRunObservation{ID: v.GetId(), WorkspaceID: v.GetWorkspaceId(), AutomationID: v.GetAutomationId(), TriggerID: v.GetTriggerId(), TriggerType: v.GetTriggerType(), TaskID: v.GetTaskId(), SessionID: v.GetSessionId(), TurnID: v.GetTurnId(), Status: v.GetStatus(), Summary: v.GetSummary(), ErrorMessage: v.GetErrorMessage(), RetryGroupID: v.GetRetryGroupId(), RetryParentRunID: v.GetRetryParentRunId(), AttemptNumber: v.GetAttemptNumber(), CreatedAt: v.GetCreatedAt(), ResourceVersion: v.GetResourceVersion(), ObservedAt: v.GetObservedAt()}
}
