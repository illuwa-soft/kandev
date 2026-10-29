package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kandev/kandev/internal/automation"
	taskmodels "github.com/kandev/kandev/internal/task/models"
	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func automationReadFixture(t *testing.T) (*pluginHost, *automation.Store, *automation.Automation) {
	t.Helper()
	db, err := sqlx.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	source, err := automation.NewStore(db, db)
	require.NoError(t, err)
	a := &automation.Automation{ID: "automation-one", WorkspaceID: "workspace-one", Name: "Morning check", Enabled: true, Prompt: "private prompt", WebhookSecret: "private-webhook-secret"}
	require.NoError(t, source.CreateAutomation(context.Background(), a))
	require.NoError(t, source.CreateTrigger(context.Background(), &automation.AutomationTrigger{ID: "schedule", AutomationID: a.ID, Type: automation.TriggerTypeScheduled, Config: json.RawMessage(`{"cron_expression":"0 9 * * *","timezone":"UTC","secret":"never-export"}`), Enabled: true}))
	base := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	for i, id := range []string{"run-one", "run-two"} {
		r := &automation.AutomationRun{ID: id, AutomationID: a.ID, TriggerID: "schedule", TriggerType: automation.TriggerTypeScheduled, Status: automation.RunStatusTriggered, CreatedAt: base.Add(time.Duration(i) * time.Minute), TriggerData: json.RawMessage(`{"password":"never-export"}`), RetryGroupID: "retry-group", AttemptNumber: int64(i + 1)}
		require.NoError(t, source.CreateRun(context.Background(), r))
	}
	foreign := &automation.Automation{ID: "automation-foreign", WorkspaceID: "workspace-two", Name: "Foreign"}
	require.NoError(t, source.CreateAutomation(context.Background(), foreign))
	require.NoError(t, source.CreateRun(context.Background(), &automation.AutomationRun{ID: "foreign-run", AutomationID: foreign.ID, Status: automation.RunStatusSkipped}))
	data := &fakeTaskDataSource{workspaces: []*taskmodels.Workspace{{ID: "workspace-one"}, {ID: "workspace-two"}}}
	h := newExactReadHost(t, data, []string{"automations"}, []string{"host.v2.read:automations"})
	h.service.SetAutomationHistorySource(automation.NewService(source, nil, testLogger(t)))
	return h, source, a
}

// @covers AC-PLUGINS-AUTOMATION-READS-001.1 AC-PLUGINS-AUTOMATION-READS-001.4 AC-PLUGINS-AUTOMATION-READS-001.7
func TestAutomationHistorySourceAndWireSnapshot(t *testing.T) {
	h, source, a := automationReadFixture(t)
	reader, ok := pluginsdk.AutomationHistory(dialPluginHostOverWire(t, h))
	require.True(t, ok)
	q := pluginsdk.AutomationHistoryQuery{RequestID: "first", WorkspaceID: "workspace-one", Page: pluginsdk.ExactReadPage{Limit: 1}}
	definitions, err := reader.ListAutomations(context.Background(), q)
	require.NoError(t, err)
	require.Len(t, definitions.Items, 1)
	require.Equal(t, a.ID, definitions.Items[0].ID)
	require.Equal(t, "UTC", definitions.Items[0].Schedules[0].Timezone)
	encoded, err := json.Marshal(definitions)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "private")
	require.NotContains(t, string(encoded), "never-export")
	page, err := reader.ListRuns(context.Background(), q)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.True(t, page.PageInfo.HasMore)
	require.NotEmpty(t, page.PageInfo.Receipt.InstallationID)
	q.Page = pluginsdk.ExactReadPage{Limit: 1, Cursor: page.PageInfo.NextCursor, SnapshotVersion: page.PageInfo.SnapshotVersion}
	next, err := reader.ListRuns(context.Background(), q)
	require.NoError(t, err)
	require.Len(t, next.Items, 1)
	require.False(t, next.PageInfo.HasMore)
	require.NotEqual(t, page.Items[0].ID, next.Items[0].ID)
	require.Empty(t, next.Items[0].TaskID)
	require.Equal(t, "retry-group", next.Items[0].RetryGroupID)
	before, err := source.GetRun(context.Background(), next.Items[0].ID)
	require.NoError(t, err)
	_, _, err = reader.GetRun(context.Background(), pluginsdk.AutomationRunLookup{RequestID: "known", WorkspaceID: "workspace-one", RunID: before.ID})
	require.NoError(t, err)
	after, err := source.GetRun(context.Background(), before.ID)
	require.NoError(t, err)
	require.Equal(t, before, after)
	encoded, err = json.Marshal(next.Items)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "never-export")
}

// @covers AC-PLUGINS-AUTOMATION-READS-001.6 AC-PLUGINS-AUTOMATION-READS-001.9
func TestAutomationHistoryDeniesBeforeSourceAndRejectsForeignIdentity(t *testing.T) {
	bare := &pluginHost{}
	_, err := bare.ListAutomationRuns(context.Background(), pluginsdk.AutomationHistoryQuery{RequestID: "read", WorkspaceID: "workspace-one"})
	require.Equal(t, codes.Unavailable, status.Code(err))
	h, _, _ := automationReadFixture(t)
	_, _, err = h.GetAutomationRun(context.Background(), pluginsdk.AutomationRunLookup{RequestID: "foreign", WorkspaceID: "workspace-one", RunID: "foreign-run"})
	require.Equal(t, codes.NotFound, status.Code(err))
	_, _, err = h.GetAutomationRun(context.Background(), pluginsdk.AutomationRunLookup{RequestID: "missing", WorkspaceID: "workspace-one", RunID: "missing"})
	require.Equal(t, codes.NotFound, status.Code(err))
	_, err = h.ListAutomationRuns(context.Background(), pluginsdk.AutomationHistoryQuery{RequestID: "scope", WorkspaceID: "workspace-two"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	page, err := h.ListAutomationRuns(context.Background(), pluginsdk.AutomationHistoryQuery{RequestID: "page", WorkspaceID: "workspace-one", Page: pluginsdk.ExactReadPage{Limit: 1}})
	require.NoError(t, err)
	_, err = h.ListAutomationRuns(context.Background(), pluginsdk.AutomationHistoryQuery{RequestID: "filter", WorkspaceID: "workspace-one", AutomationID: "changed", Page: pluginsdk.ExactReadPage{Limit: 1, Cursor: page.PageInfo.NextCursor, SnapshotVersion: page.PageInfo.SnapshotVersion}})
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
}

// @covers AC-PLUGINS-AUTOMATION-READS-001.9
func TestAutomationHistoryCapabilityDiscoveryTracksLateSourceWiring(t *testing.T) {
	h, _, _ := automationReadFixture(t)
	for _, available := range []bool{true, false} {
		if !available {
			h.service.SetAutomationHistorySource(nil)
		}
		capability, err := h.GetCapabilityContext(context.Background(), "workspace-one")
		require.NoError(t, err)
		found := 0
		for _, operation := range capability.Operations {
			switch operation.Method {
			case "ListAutomationDefinitionsExact", "ListAutomationRunsExact", "GetAutomationRunExact":
				found++
				require.Equal(t, "host.v2.read:automations", operation.CapabilityID)
				require.Equal(t, available, operation.Supported)
				require.Equal(t, available, operation.Authorized)
				if !available {
					require.Equal(t, "automation_history_unavailable", operation.UnavailableReason)
				}
			}
		}
		require.Equal(t, 3, found)
	}
}

// @covers AC-PLUGINS-AUTOMATION-READS-001.4 AC-PLUGINS-AUTOMATION-READS-001.6
func TestAutomationHistoryRevocationAndExpiredSnapshots(t *testing.T) {
	h, _, _ := automationReadFixture(t)
	query := pluginsdk.AutomationHistoryQuery{RequestID: "start", WorkspaceID: "workspace-one", Page: pluginsdk.ExactReadPage{Limit: 1}}
	page, err := h.ListAutomationRuns(context.Background(), query)
	require.NoError(t, err)
	query.Page = pluginsdk.ExactReadPage{Cursor: page.PageInfo.NextCursor, SnapshotVersion: page.PageInfo.SnapshotVersion}
	query.Page.Cursor = "malformed"
	_, err = h.ListAutomationRuns(context.Background(), query)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	query.Page.Cursor = page.PageInfo.NextCursor
	h.exactSnapshots.mu.Lock()
	h.exactSnapshots.snapshots[page.PageInfo.SnapshotVersion].createdAt = time.Now().Add(-2 * exactSnapshotTTL)
	h.exactSnapshots.mu.Unlock()
	_, err = h.ListAutomationRuns(context.Background(), query)
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
	query.Page = pluginsdk.ExactReadPage{Limit: 1}
	page, err = h.ListAutomationRuns(context.Background(), query)
	require.NoError(t, err)
	query.Page = pluginsdk.ExactReadPage{Cursor: page.PageInfo.NextCursor, SnapshotVersion: page.PageInfo.SnapshotVersion}
	_, err = h.service.approvalRevoke(h.installationID, query.WorkspaceID, "human", "revoke", "revoke-history")
	require.NoError(t, err)
	_, err = h.ListAutomationRuns(context.Background(), query)
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	_, _, err = h.GetAutomationRun(context.Background(), pluginsdk.AutomationRunLookup{RequestID: "revoked", WorkspaceID: query.WorkspaceID, RunID: "run-one"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}

type failingAutomationHistorySource struct {
	automationHistorySource
	calls int
}

func (s *failingAutomationHistorySource) ListAutomations(context.Context, string) ([]*automation.Automation, error) {
	s.calls++
	return nil, errors.New("private upstream credentials")
}

// @covers AC-PLUGINS-AUTOMATION-READS-001.6 AC-PLUGINS-AUTOMATION-READS-001.7
func TestAutomationHistoryAuthorizesBeforeSourceAndSanitizesReadFailure(t *testing.T) {
	h, _, _ := automationReadFixture(t)
	source := &failingAutomationHistorySource{}
	h.service.SetAutomationHistorySource(source)
	query := pluginsdk.AutomationHistoryQuery{RequestID: "denied", WorkspaceID: "workspace-two"}
	_, err := h.ListAutomationRuns(context.Background(), query)
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	require.Zero(t, source.calls)
	query.WorkspaceID = "workspace-one"
	_, err = h.ListAutomationRuns(context.Background(), query)
	require.Equal(t, codes.Unavailable, status.Code(err))
	require.Equal(t, 1, source.calls)
	require.False(t, strings.Contains(err.Error(), "private"))
}

// @covers AC-PLUGINS-AUTOMATION-READS-001.2 AC-PLUGINS-AUTOMATION-READS-001.4
func TestAutomationHistoryExhaustsServicePagesIncludingTasklessRuns(t *testing.T) {
	h, source, definition := automationReadFixture(t)
	ctx := t.Context()
	for index := 0; index < 205; index++ {
		require.NoError(t, source.CreateRun(ctx, &automation.AutomationRun{
			ID: fmt.Sprintf("paged-%03d", index), AutomationID: definition.ID,
			Status: automation.RunStatusSkipped, TriggerType: automation.TriggerTypeScheduled,
			CreatedAt: time.Date(2026, 10, 2, 9, 0, 0, index, time.UTC),
		}))
	}
	query := pluginsdk.AutomationHistoryQuery{RequestID: "exhaust", WorkspaceID: definition.WorkspaceID, AutomationID: definition.ID, Page: pluginsdk.ExactReadPage{Limit: 130}}
	seen := make(map[string]bool)
	for {
		page, err := h.ListAutomationRuns(ctx, query)
		require.NoError(t, err)
		for _, run := range page.Items {
			require.False(t, seen[run.ID], "repeated run %s", run.ID)
			seen[run.ID] = true
			if strings.HasPrefix(run.ID, "paged-") {
				require.Empty(t, run.TaskID)
				require.Equal(t, "skipped", run.Status)
			}
		}
		if !page.PageInfo.HasMore {
			break
		}
		query.Page.Cursor, query.Page.SnapshotVersion = page.PageInfo.NextCursor, page.PageInfo.SnapshotVersion
	}
	require.Len(t, seen, 207)
}

// @covers AC-PLUGINS-AUTOMATION-READS-001.3 AC-PLUGINS-AUTOMATION-READS-001.5
func TestAutomationHistoryRefreshesOutcomeWithoutChangingEarlierSnapshot(t *testing.T) {
	h, source, definition := automationReadFixture(t)
	ctx := t.Context()
	run := &automation.AutomationRun{ID: "live", AutomationID: definition.ID, Status: automation.RunStatusTriggered, TriggerType: automation.TriggerTypeManual, CreatedAt: time.Date(2026, 10, 3, 9, 0, 0, 0, time.UTC)}
	require.NoError(t, source.CreateRun(ctx, run))
	query := pluginsdk.AutomationHistoryQuery{RequestID: "before", WorkspaceID: definition.WorkspaceID, Page: pluginsdk.ExactReadPage{Limit: 1}}
	before, err := h.ListAutomationRuns(ctx, query)
	require.NoError(t, err)
	require.Equal(t, run.ID, before.Items[0].ID)
	require.Equal(t, "triggered", before.Items[0].Status)
	require.NoError(t, source.MarkRunTerminal(ctx, run.ID, "", "", automation.RunStatusFailed, "Provider unavailable"))
	current, receipt, err := h.GetAutomationRun(ctx, pluginsdk.AutomationRunLookup{RequestID: "after", WorkspaceID: definition.WorkspaceID, RunID: run.ID})
	require.NoError(t, err)
	require.Equal(t, "failed", current.Status)
	require.Equal(t, "Provider unavailable", current.ErrorMessage)
	require.NotEqual(t, before.Items[0].ResourceVersion, current.ResourceVersion)
	require.Equal(t, before.Items[0].CreatedAt, current.CreatedAt)
	require.Equal(t, "after", receipt.RequestID)
	query.Page.SnapshotVersion = before.PageInfo.SnapshotVersion
	repeated, err := h.ListAutomationRuns(ctx, query)
	require.NoError(t, err)
	require.Equal(t, before.Items, repeated.Items)
}
