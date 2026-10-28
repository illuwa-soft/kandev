package changes

import (
	"context"
	"testing"
	"time"

	"github.com/kandev/kandev/internal/common/turnchanges"
	"github.com/kandev/kandev/internal/task/models"
)

type blockingExportCandidates struct {
	*coordinatorStore
	started chan struct{}
	release chan struct{}
}

func (s *blockingExportCandidates) ListTurnChangeExportCandidates(
	context.Context, time.Time, int,
) ([]*models.TurnRepositoryChangeSet, error) {
	close(s.started)
	<-s.release
	return nil, nil
}

func TestExportProcessorStopWaitsForInflightPass(t *testing.T) {
	store := newCoordinatorStore()
	blocking := &blockingExportCandidates{
		coordinatorStore: store, started: make(chan struct{}), release: make(chan struct{}),
	}
	coordinator := NewCoordinator(blocking, coordinatorTurnReaderMap{}, &coordinatorPolicy{
		result: turnchanges.CapturePolicy{},
	}, NewContentService(blocking, nil), nil)
	processor := NewExportProcessor(blocking, coordinator)
	processor.SetClientProvider(staticExportProvider{ok: true})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	processor.Start(ctx)
	select {
	case <-blocking.started:
	case <-time.After(time.Second):
		t.Fatal("processor did not start a pass")
	}

	stopped := make(chan struct{})
	go func() {
		processor.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
		t.Fatal("stop returned before the in-flight pass drained")
	case <-time.After(50 * time.Millisecond):
	}
	close(blocking.release)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("stop did not join the drained worker")
	}

	processor.Stop()
}
