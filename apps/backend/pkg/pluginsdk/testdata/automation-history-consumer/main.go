package main

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This independent executable consumes only public SDK types and transport.
type consumer struct {
	pluginsdk.UnimplementedPlugin
	once  sync.Once
	ready chan struct{}
}

func (p *consumer) SetHost(host pluginsdk.Host) {
	p.UnimplementedPlugin.SetHost(host)
	p.once.Do(func() { close(p.ready) })
}

func (p *consumer) HandleWebhook(ctx context.Context, request *pluginsdk.WebhookRequest) (*pluginsdk.WebhookResponse, error) {
	select {
	case <-p.ready:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	var query pluginsdk.AutomationRunQuery
	if err := json.Unmarshal(request.Body, &query); err != nil {
		return &pluginsdk.WebhookResponse{Status: 400}, nil
	}
	reader, ok := pluginsdk.HostAutomationHistory(p.Host())
	if !ok {
		return &pluginsdk.WebhookResponse{Status: 501}, nil
	}
	page, err := reader.ListRuns(ctx, query)
	if err != nil {
		responseStatus := int32(502)
		switch status.Code(err) {
		case codes.Unimplemented:
			responseStatus = 501
		case codes.PermissionDenied:
			responseStatus = 403
		case codes.FailedPrecondition, codes.InvalidArgument:
			responseStatus = 409
		}
		return &pluginsdk.WebhookResponse{Status: responseStatus}, nil
	}
	body, err := json.Marshal(page)
	if err != nil {
		return nil, err
	}
	return &pluginsdk.WebhookResponse{Status: 200, Body: body}, nil
}

func main() {
	pluginsdk.Serve(&consumer{ready: make(chan struct{})})
}
