package pluginsdk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type automationHistoryHost struct {
	*dataRecordingHost
	reader AutomationHistoryReader
}

func (h *automationHistoryHost) AutomationHistory() AutomationHistoryReader { return h.reader }

type automationHistoryRecorder struct{ query AutomationHistoryQuery }

func (r *automationHistoryRecorder) ListAutomations(_ context.Context, q AutomationHistoryQuery) (AutomationDefinitionPage, error) {
	r.query = q
	return AutomationDefinitionPage{Items: []AutomationDefinition{{ID: "a", WorkspaceID: q.WorkspaceID, Name: "Release check", Schedules: []AutomationSchedule{{TriggerID: "trigger", Expression: "0 9 * * *", Timezone: "UTC", Enabled: true}}}}}, nil
}

func (r *automationHistoryRecorder) ListRuns(_ context.Context, q AutomationHistoryQuery) (AutomationRunPage, error) {
	r.query = q
	return AutomationRunPage{Items: []AutomationRunObservation{{ID: "run", AutomationID: "a", WorkspaceID: q.WorkspaceID, Status: "skipped", RetryParentRunID: "parent", AttemptNumber: 7}}, PageInfo: ExactReadPageInfo{NextCursor: "opaque", HasMore: true, SnapshotVersion: "snapshot"}}, nil
}

func (r *automationHistoryRecorder) GetRun(_ context.Context, q AutomationRunLookup) (AutomationRunObservation, HostReadReceipt, error) {
	return AutomationRunObservation{ID: q.RunID, WorkspaceID: q.WorkspaceID, SessionID: "s", TurnID: "t", Status: "completed"}, HostReadReceipt{WorkspaceID: q.WorkspaceID, RequestID: q.RequestID}, nil
}

// @covers AC-PLUGINS-AUTOMATION-READS-001.1 AC-PLUGINS-AUTOMATION-READS-001.2
func TestAutomationHistoryWireRoundTrip(t *testing.T) {
	r := &automationHistoryRecorder{}
	host := dialHostOverBufconn(t, &automationHistoryHost{dataRecordingHost: &dataRecordingHost{}, reader: r})
	reader, ok := AutomationHistory(host)
	require.True(t, ok)
	q := AutomationHistoryQuery{RequestID: "request", WorkspaceID: "workspace", AutomationID: "a", Page: ExactReadPage{Limit: 1, Cursor: "before", SnapshotVersion: "old"}}
	definitions, err := reader.ListAutomations(context.Background(), q)
	require.NoError(t, err)
	require.Equal(t, q, r.query)
	require.Equal(t, "UTC", definitions.Items[0].Schedules[0].Timezone)
	runs, err := reader.ListRuns(context.Background(), q)
	require.NoError(t, err)
	require.Equal(t, q, r.query)
	require.Equal(t, "skipped", runs.Items[0].Status)
	require.Empty(t, runs.Items[0].TaskID)
	require.Equal(t, int64(7), runs.Items[0].AttemptNumber)
	require.Equal(t, "opaque", runs.PageInfo.NextCursor)
	run, receipt, err := reader.GetRun(context.Background(), AutomationRunLookup{RequestID: "read", WorkspaceID: "workspace", RunID: "run"})
	require.NoError(t, err)
	require.Equal(t, "t", run.TurnID)
	require.Equal(t, "read", receipt.RequestID)
}

// @covers AC-PLUGINS-AUTOMATION-READS-001.9
func TestAutomationHistoryOldHostIsUnsupported(t *testing.T) {
	host := dialHostOverBufconn(t, &dataRecordingHost{})
	reader, ok := AutomationHistory(host)
	require.True(t, ok)
	_, err := reader.ListRuns(context.Background(), AutomationHistoryQuery{WorkspaceID: "workspace", RequestID: "request"})
	require.Equal(t, codes.Unimplemented, status.Code(err))
	_, ok = AutomationHistory(&dataRecordingHost{})
	require.False(t, ok)
}
