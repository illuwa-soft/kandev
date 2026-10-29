package automation

import (
	"testing"
	"time"

	"github.com/kandev/kandev/internal/testutil"
	"github.com/stretchr/testify/require"
)

// @covers AC-PLUGINS-AUTOMATION-READS-001.3 AC-PLUGINS-AUTOMATION-READS-001.5
func TestRunObservationUsesCurrentListProjection(t *testing.T) {
	for _, dialect := range []string{"sqlite", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			var store *Store
			if dialect == "sqlite" {
				store = setupTestStore(t)
			} else {
				database := testutil.OpenIsolatedPostgres(t, testutil.PostgresDSNFromEnv(t))
				var err error
				store, err = NewStore(database, database)
				require.NoError(t, err)
			}
			require.NoError(t, createHistoryReadTables(store))
			service := NewService(store, nil, nil)
			ctx := t.Context()
			require.NoError(t, store.CreateAutomation(ctx, &Automation{ID: "projection", WorkspaceID: "workspace", Name: "Projection"}))
			_, err := store.db.Exec(`INSERT INTO tasks (id, archived_at) VALUES ('archived', CURRENT_TIMESTAMP)`)
			require.NoError(t, err)
			_, err = store.db.Exec(`INSERT INTO task_session_messages (id, task_id, turn_id, author_type, content, type, created_at) VALUES ('message', 'archived', 'turn', 'agent', 'Exact turn result', 'message', CURRENT_TIMESTAMP)`)
			require.NoError(t, err)
			for _, taskID := range []string{"archived", "deleted"} {
				run := &AutomationRun{ID: "run-" + taskID, AutomationID: "projection", TaskID: taskID, SessionID: "session", TurnID: "turn", Status: RunStatusTaskCreated, CreatedAt: time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)}
				require.NoError(t, store.CreateRun(ctx, run))
				before, err := store.GetRun(ctx, run.ID)
				require.NoError(t, err)
				observed, err := service.GetRunObservation(ctx, run.ID)
				require.NoError(t, err)
				wantStatus := RunStatusArchived
				if taskID == "deleted" {
					wantStatus = RunStatusCancelled
					require.Empty(t, observed.TaskID)
				}
				require.Equal(t, wantStatus, observed.Status)
				require.Equal(t, "Exact turn result", observed.Summary)
				after, err := store.GetRun(ctx, run.ID)
				require.NoError(t, err)
				require.Equal(t, before, after)
			}
			missing, err := service.GetRunObservation(ctx, "missing")
			require.NoError(t, err)
			require.Nil(t, missing)
		})
	}
}

func createHistoryReadTables(store *Store) error {
	_, err := store.db.Exec(`
		CREATE TABLE tasks (id TEXT PRIMARY KEY, archived_at TIMESTAMPTZ);
		CREATE TABLE task_sessions (id TEXT PRIMARY KEY, task_id TEXT NOT NULL, is_primary INTEGER NOT NULL DEFAULT 0, state TEXT NOT NULL DEFAULT 'CREATED');
		CREATE TABLE task_session_messages (id TEXT PRIMARY KEY, task_id TEXT NOT NULL, turn_id TEXT NOT NULL, author_type TEXT NOT NULL, content TEXT NOT NULL, type TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL)`)
	return err
}
