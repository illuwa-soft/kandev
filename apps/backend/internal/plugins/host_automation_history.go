package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/kandev/kandev/internal/automation"
	"github.com/kandev/kandev/pkg/pluginsdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	exactReadListAutomationDefinitions = "ListAutomationDefinitionsExact"
	exactReadListAutomationRuns        = "ListAutomationRunsExact"
	exactReadGetAutomationRun          = "GetAutomationRunExact"
	automationHistoryCapability        = "host.v2.read:automations"
	automationHistoryUnavailable       = "automation_history_unavailable"
)

type automationHistorySource interface {
	ListAutomations(context.Context, string) ([]*automation.Automation, error)
	GetAutomation(context.Context, string) (*automation.Automation, error)
	ListRunPage(context.Context, string, string, int) (*automation.AutomationRunsPage, error)
	GetRunObservation(context.Context, string) (*automation.AutomationRun, error)
}

// SetAutomationHistorySource connects read-only automation data to existing and future plugin Hosts.
func (s *Service) SetAutomationHistorySource(source automationHistorySource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.automationHistory = source
}

func (h *pluginHost) AutomationHistory() pluginsdk.AutomationHistoryReader {
	return automationHistoryReader{host: h}
}

type automationHistoryReader struct{ host *pluginHost }

func (r automationHistoryReader) ListAutomations(ctx context.Context, q pluginsdk.AutomationHistoryQuery) (pluginsdk.AutomationDefinitionPage, error) {
	return r.host.ListAutomationDefinitions(ctx, q)
}

func (r automationHistoryReader) ListRuns(ctx context.Context, q pluginsdk.AutomationHistoryQuery) (pluginsdk.AutomationRunPage, error) {
	return r.host.ListAutomationRuns(ctx, q)
}

func (r automationHistoryReader) GetRun(ctx context.Context, q pluginsdk.AutomationRunLookup) (pluginsdk.AutomationRunObservation, pluginsdk.HostReadReceipt, error) {
	return r.host.GetAutomationRun(ctx, q)
}

func (h *pluginHost) automationHistorySource() automationHistorySource {
	if h.service == nil {
		return nil
	}
	h.service.mu.Lock()
	defer h.service.mu.Unlock()
	return h.service.automationHistory
}

func (h *pluginHost) ListAutomationDefinitions(ctx context.Context, q pluginsdk.AutomationHistoryQuery) (pluginsdk.AutomationDefinitionPage, error) {
	identity, unlock, err := h.exactReadAdmission(ctx, q.RequestID, q.WorkspaceID, exactReadListAutomationDefinitions, automationHistoryCapability, q.AutomationID)
	if err != nil {
		return pluginsdk.AutomationDefinitionPage{}, err
	}
	defer unlock()
	source := h.automationHistorySource()
	if source == nil {
		return pluginsdk.AutomationDefinitionPage{}, status.Error(codes.Unavailable, automationHistoryUnavailable)
	}
	items, info, err := exactPage(h, identity, q.RequestID, q.Page, func() ([]pluginsdk.AutomationDefinition, error) {
		models, err := readAutomationDefinitions(ctx, source, q)
		if err != nil {
			return nil, err
		}
		out := make([]pluginsdk.AutomationDefinition, len(models))
		observed := time.Now().UTC().Format(time.RFC3339Nano)
		for i, model := range models {
			out[i], err = automationDefinitionObservation(model, observed)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	})
	return pluginsdk.AutomationDefinitionPage{Items: items, PageInfo: info}, err
}

func (h *pluginHost) ListAutomationRuns(ctx context.Context, q pluginsdk.AutomationHistoryQuery) (pluginsdk.AutomationRunPage, error) {
	identity, unlock, err := h.exactReadAdmission(ctx, q.RequestID, q.WorkspaceID, exactReadListAutomationRuns, automationHistoryCapability, q.AutomationID)
	if err != nil {
		return pluginsdk.AutomationRunPage{}, err
	}
	defer unlock()
	source := h.automationHistorySource()
	if source == nil {
		return pluginsdk.AutomationRunPage{}, status.Error(codes.Unavailable, automationHistoryUnavailable)
	}
	items, info, err := exactPage(h, identity, q.RequestID, q.Page, func() ([]pluginsdk.AutomationRunObservation, error) {
		models, err := readAutomationDefinitions(ctx, source, q)
		if err != nil {
			return nil, err
		}
		return readAutomationRuns(ctx, source, q.WorkspaceID, models)
	})
	return pluginsdk.AutomationRunPage{Items: items, PageInfo: info}, err
}

func (h *pluginHost) GetAutomationRun(ctx context.Context, q pluginsdk.AutomationRunLookup) (pluginsdk.AutomationRunObservation, pluginsdk.HostReadReceipt, error) {
	identity, unlock, err := h.exactReadAdmission(ctx, q.RequestID, q.WorkspaceID, exactReadGetAutomationRun, automationHistoryCapability, q.RunID)
	if err != nil {
		return pluginsdk.AutomationRunObservation{}, pluginsdk.HostReadReceipt{}, err
	}
	defer unlock()
	if !isBoundedApprovalIdentifier(q.RunID) {
		return pluginsdk.AutomationRunObservation{}, pluginsdk.HostReadReceipt{}, status.Error(codes.InvalidArgument, "run_id is required and bounded")
	}
	source := h.automationHistorySource()
	if source == nil {
		return pluginsdk.AutomationRunObservation{}, pluginsdk.HostReadReceipt{}, status.Error(codes.Unavailable, automationHistoryUnavailable)
	}
	run, err := source.GetRunObservation(ctx, q.RunID)
	if err != nil {
		return pluginsdk.AutomationRunObservation{}, pluginsdk.HostReadReceipt{}, automationHistoryReadError(err)
	}
	if run == nil || run.ID != q.RunID {
		return pluginsdk.AutomationRunObservation{}, pluginsdk.HostReadReceipt{}, status.Error(codes.NotFound, "automation run was not found")
	}
	definition, err := source.GetAutomation(ctx, run.AutomationID)
	if err != nil {
		return pluginsdk.AutomationRunObservation{}, pluginsdk.HostReadReceipt{}, automationHistoryReadError(err)
	}
	if definition == nil || definition.WorkspaceID != q.WorkspaceID || definition.ID != run.AutomationID {
		return pluginsdk.AutomationRunObservation{}, pluginsdk.HostReadReceipt{}, status.Error(codes.NotFound, "automation run was not found")
	}
	receipt := singleReadReceipt(identity, q.RequestID)
	return automationRunObservation(run, q.WorkspaceID, receipt.ObservedAt), receipt, nil
}

func readAutomationDefinitions(ctx context.Context, source automationHistorySource, q pluginsdk.AutomationHistoryQuery) ([]*automation.Automation, error) {
	if q.AutomationID != "" && !isBoundedApprovalIdentifier(q.AutomationID) {
		return nil, status.Error(codes.InvalidArgument, "automation_id must be bounded")
	}
	models, err := source.ListAutomations(ctx, q.WorkspaceID)
	if err != nil {
		return nil, automationHistoryReadError(err)
	}
	out := make([]*automation.Automation, 0, len(models))
	for _, model := range models {
		if model == nil {
			return nil, status.Error(codes.Unavailable, "automation projection is incomplete")
		}
		if model.WorkspaceID == q.WorkspaceID && (q.AutomationID == "" || model.ID == q.AutomationID) {
			out = append(out, model)
		}
	}
	if q.AutomationID != "" && len(out) == 0 {
		return nil, status.Error(codes.NotFound, "automation was not found")
	}
	if len(out) > exactSnapshotMaxRows {
		return nil, status.Error(codes.ResourceExhausted, "automation snapshot exceeds its bounded size")
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func readAutomationRuns(ctx context.Context, source automationHistorySource, workspaceID string, definitions []*automation.Automation) ([]pluginsdk.AutomationRunObservation, error) {
	models := make([]*automation.AutomationRun, 0)
	for _, definition := range definitions {
		runs, err := readAutomationRunPages(ctx, source, definition.ID, exactSnapshotMaxRows-len(models))
		if err != nil {
			return nil, err
		}
		models = append(models, runs...)
	}
	sort.Slice(models, func(i, j int) bool {
		if models[i].CreatedAt.Equal(models[j].CreatedAt) {
			return models[i].ID < models[j].ID
		}
		return models[i].CreatedAt.After(models[j].CreatedAt)
	})
	out := make([]pluginsdk.AutomationRunObservation, len(models))
	observed := time.Now().UTC().Format(time.RFC3339Nano)
	for i, model := range models {
		out[i] = automationRunObservation(model, workspaceID, observed)
	}
	return out, nil
}

func readAutomationRunPages(ctx context.Context, source automationHistorySource, automationID string, remaining int) ([]*automation.AutomationRun, error) {
	out := make([]*automation.AutomationRun, 0)
	seen := make(map[string]bool)
	cursor := ""
	for {
		if err := ctx.Err(); err != nil {
			return nil, status.FromContextError(err).Err()
		}
		page, err := source.ListRunPage(ctx, automationID, cursor, 200)
		if err != nil {
			return nil, automationHistoryReadError(err)
		}
		if page == nil {
			return nil, status.Error(codes.Unavailable, "automation run projection is incomplete")
		}
		for _, run := range page.Items {
			if run == nil || run.AutomationID != automationID {
				return nil, status.Error(codes.Unavailable, "automation run projection is incomplete")
			}
			out = append(out, run)
		}
		if len(out) > remaining {
			return nil, status.Error(codes.ResourceExhausted, "automation run snapshot exceeds its bounded size")
		}
		if page.NextCursor == "" {
			return out, nil
		}
		if seen[page.NextCursor] || len(page.Items) == 0 {
			return nil, status.Error(codes.Unavailable, "automation run paging did not advance")
		}
		seen[page.NextCursor] = true
		cursor = page.NextCursor
	}
}

func automationDefinitionObservation(model *automation.Automation, observed string) (pluginsdk.AutomationDefinition, error) {
	item := pluginsdk.AutomationDefinition{
		ID: model.ID, WorkspaceID: model.WorkspaceID, Name: model.Name, Enabled: model.Enabled,
		AgentProfileID: model.AgentProfileID, ExecutorProfileID: model.ExecutorProfileID,
		CreatedAt: automationSourceTime(model.CreatedAt), UpdatedAt: automationSourceTime(model.UpdatedAt),
		Schedules: make([]pluginsdk.AutomationSchedule, 0, len(model.Triggers)),
	}
	for _, trigger := range model.Triggers {
		schedule := pluginsdk.AutomationSchedule{TriggerID: trigger.ID, TriggerType: string(trigger.Type), Enabled: trigger.Enabled}
		if trigger.Type == automation.TriggerTypeScheduled {
			var config automation.ScheduledTriggerConfig
			if err := json.Unmarshal(trigger.Config, &config); err != nil {
				return pluginsdk.AutomationDefinition{}, status.Error(codes.Unavailable, "automation schedule projection is invalid")
			}
			schedule.Expression, schedule.Timezone = config.CronExpression, config.Timezone
		}
		item.Schedules = append(item.Schedules, schedule)
	}
	sort.Slice(item.Schedules, func(i, j int) bool { return item.Schedules[i].TriggerID < item.Schedules[j].TriggerID })
	item.ResourceVersion = exactFilterDigest(item)
	item.ObservedAt = observed
	return item, nil
}

func automationRunObservation(model *automation.AutomationRun, workspaceID, observed string) pluginsdk.AutomationRunObservation {
	item := pluginsdk.AutomationRunObservation{
		ID: model.ID, WorkspaceID: workspaceID, AutomationID: model.AutomationID,
		TriggerID: model.TriggerID, TriggerType: string(model.TriggerType),
		TaskID: model.TaskID, SessionID: model.SessionID, TurnID: model.TurnID,
		Status: string(model.Status), Summary: model.Summary, ErrorMessage: model.ErrorMessage,
		RetryGroupID: model.RetryGroupID, RetryParentRunID: model.RetryParentRunID, AttemptNumber: model.AttemptNumber,
		CreatedAt: automationSourceTime(model.CreatedAt),
	}
	item.ResourceVersion = exactFilterDigest(item)
	item.ObservedAt = observed
	return item
}

func automationSourceTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func automationHistoryReadError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return status.FromContextError(err).Err()
	}
	if errors.Is(err, automation.ErrAutomationNotFound) {
		return status.Error(codes.NotFound, "automation run was not found")
	}
	return status.Error(codes.Unavailable, "automation history read failed")
}
