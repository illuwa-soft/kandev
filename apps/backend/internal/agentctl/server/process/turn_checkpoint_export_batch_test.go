package process

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kandev/kandev/internal/common/turnchanges"
)

func TestTurnCheckpointExportBatchResumesAtCursor(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	t.Cleanup(cleanup)
	operator := NewGitOperator(repoDir, newTestLogger(t), nil)
	changeSetID, checkoutID := uuid.NewString(), uuid.NewString()
	captureHistoryEndpoint(t, operator, changeSetID, checkoutID, turnchanges.CheckpointStart)
	writeFile(t, repoDir, "a.txt", "a\n")
	writeFile(t, repoDir, "b.txt", "b\n")
	writeFile(t, repoDir, "c.txt", "c\n")
	captureHistoryEndpoint(t, operator, changeSetID, checkoutID, turnchanges.CheckpointEnd)
	request := exportRequestForPair(acceptedTurnCheckpointPair(t, operator, changeSetID, checkoutID, ""))
	request.ExportBatchVersion = turnchanges.ExportBatchVersion1
	request.MaxEntries = 2

	first, err := operator.ExportTurnCheckpoint(context.Background(), request)
	if err != nil {
		t.Fatalf("first export page: %v", err)
	}
	if len(first.Files) != 2 || first.NextCursor != 2 || first.Complete || first.TotalFiles != 3 {
		t.Fatalf("first page = files:%d cursor:%d complete:%t total:%d", len(first.Files), first.NextCursor, first.Complete, first.TotalFiles)
	}
	if first.InventoryFingerprint == "" {
		t.Fatal("batch export did not advertise an inventory fingerprint")
	}
	request.Cursor = first.NextCursor
	request.InventoryFingerprint = first.InventoryFingerprint
	second, err := operator.ExportTurnCheckpoint(context.Background(), request)
	if err != nil {
		t.Fatalf("second export page: %v", err)
	}
	if len(second.Files) != 1 || second.NextCursor != 3 || !second.Complete {
		t.Fatalf("second page = files:%d cursor:%d complete:%t", len(second.Files), second.NextCursor, second.Complete)
	}
	if !strings.HasPrefix(second.InventoryFingerprint, first.InventoryFingerprint[:8]) {
		t.Fatalf("fingerprint changed between pages: %q vs %q", first.InventoryFingerprint, second.InventoryFingerprint)
	}
}

func TestTurnCheckpointExportBatchRejectsChangedInventory(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	t.Cleanup(cleanup)
	operator := NewGitOperator(repoDir, newTestLogger(t), nil)
	changeSetID, checkoutID := uuid.NewString(), uuid.NewString()
	captureHistoryEndpoint(t, operator, changeSetID, checkoutID, turnchanges.CheckpointStart)
	writeFile(t, repoDir, "a.txt", "a\n")
	captureHistoryEndpoint(t, operator, changeSetID, checkoutID, turnchanges.CheckpointEnd)
	request := exportRequestForPair(acceptedTurnCheckpointPair(t, operator, changeSetID, checkoutID, ""))
	request.ExportBatchVersion = turnchanges.ExportBatchVersion1
	request.InventoryFingerprint = strings.Repeat("0", 64)
	_, err := operator.ExportTurnCheckpoint(context.Background(), request)
	var checkpointErr *TurnCheckpointError
	if err == nil || !errors.As(err, &checkpointErr) || checkpointErr.Reason != turnchanges.ReasonHistoryContextChanged {
		t.Fatalf("changed inventory error = %v", err)
	}
}
