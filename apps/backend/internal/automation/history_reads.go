package automation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/kandev/kandev/internal/db"
)

// GetRunObservation uses the same authorization and current projection as run history.
func (s *Service) GetRunObservation(ctx context.Context, id string) (*AutomationRun, error) {
	run, err := s.store.GetRunObservation(ctx, id)
	if err != nil || run == nil {
		return run, err
	}
	if err := s.authorizeAutomation(ctx, run.AutomationID); err != nil {
		return nil, err
	}
	return run, nil
}

// GetRunObservation returns the current read projection without mutating the stored run.
func (s *Store) GetRunObservation(ctx context.Context, id string) (*AutomationRun, error) {
	var run AutomationRun
	err := s.ro.GetContext(ctx, &run, s.ro.Rebind(`SELECT`+runTaskStateColumnsSQL+`
		FROM automation_runs ar LEFT JOIN tasks t ON t.id = ar.task_id WHERE ar.id = ?`),
		append(runTaskStateArgs(), id)...)
	var result *AutomationRun
	if db.IsMissingTableError(err) {
		result, err = s.GetRun(ctx, id)
	} else if err == nil {
		result = &run
	}
	if errors.Is(err, sql.ErrNoRows) || (err == nil && result == nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if result.RetryState == RetryStateCancelled {
		result.Status = RunStatusCancelled
	}
	result.TriggerData = json.RawMessage(result.TriggerDataJSON)
	if err := s.hydrateRunSummaries(ctx, []*AutomationRun{result}); err != nil {
		return nil, err
	}
	return result, nil
}
