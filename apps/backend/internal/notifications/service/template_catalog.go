package service

import (
	"context"
	"fmt"

	"github.com/kandev/kandev/internal/notifications/models"
	"github.com/kandev/kandev/internal/notifications/templates"
	taskmodels "github.com/kandev/kandev/internal/task/models"
)

type TemplateEvent struct {
	Event        string                       `json:"event"`
	Variables    []string                     `json:"variables"`
	Samples      []string                     `json:"samples"`
	DefaultTitle string                       `json:"default_title"`
	DefaultBody  string                       `json:"default_body"`
	Override     *models.NotificationTemplate `json:"override"`
}

type TemplateCatalog struct {
	Events              []TemplateEvent `json:"events"`
	PublicURLConfigured bool            `json:"public_url_configured"`
}

type TemplatePreview struct {
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	UsedFallback bool     `json:"used_fallback"`
	Diagnostics  []string `json:"diagnostics"`
}

func templateSamples(event string) []string {
	if event == EventSystemUpdateAvailable {
		return []string{"normal", "runtime_available", "runtime_succeeded", "runtime_failed", "runtime_interrupted", "runtime_summary", "missing_url"}
	}
	if event == EventOfficeInboxItem {
		return []string{"normal", "missing_url"}
	}
	return []string{"normal", "missing_session", "missing_title", "missing_url"}
}

func (s *Service) TemplateCatalog(ctx context.Context, userID, providerID string) (TemplateCatalog, error) {
	if err := s.ownedTemplateProvider(ctx, userID, providerID); err != nil {
		return TemplateCatalog{}, err
	}
	result := TemplateCatalog{Events: []TemplateEvent{}, PublicURLConfigured: s.publicOrigin() != ""}
	for _, event := range s.AvailableEvents() {
		p := samplePayload(event, "normal")
		override, err := s.repo.GetTemplate(ctx, userID, providerID, event)
		if err != nil {
			return TemplateCatalog{}, err
		}
		result.Events = append(result.Events, TemplateEvent{Event: event, Variables: templates.Variables(event), Samples: templateSamples(event), DefaultTitle: p.Title, DefaultBody: p.Body, Override: override})
	}
	return result, nil
}

// PreviewTemplate reads only provider ownership and synthetic literals, never task data or senders.
func (s *Service) PreviewTemplate(ctx context.Context, userID, providerID, event, sample string, tpl *models.NotificationTemplate) (TemplatePreview, error) {
	if err := s.ownedTemplateProvider(ctx, userID, providerID); err != nil {
		return TemplatePreview{}, err
	}
	if err := s.validateEvents([]string{event}); err != nil {
		return TemplatePreview{}, templates.ErrInvalid
	}
	if !containsEvent(templateSamples(event), sample) || tpl.Version != 1 {
		return TemplatePreview{}, templates.ErrInvalid
	}
	if err := templates.Validate(tpl.Title, tpl.Body, templates.Variables(event)); err != nil {
		return TemplatePreview{}, err
	}
	p := samplePayload(event, sample)
	synthetic := &Service{taskRepo: sampleTasks{}, publicBaseURL: s.publicOrigin()}
	if sample == "missing_url" {
		synthetic.publicBaseURL = ""
	}
	values := synthetic.templateValues(ctx, p)
	title, body, err := renderPair(tpl, event, values)
	if err != nil {
		return TemplatePreview{Title: p.Title, Body: p.Body, UsedFallback: true, Diagnostics: []string{"unavailable_or_invalid_value"}}, nil
	}
	return TemplatePreview{Title: title, Body: body, Diagnostics: []string{}}, nil
}

func samplePayload(event, sample string) notificationPayload {
	p := notificationPayload{EventType: event, Payload: map[string]string{}}
	if event == EventOfficeInboxItem {
		p.Title, p.Body = "Inbox: approval", "Example inbox item"
		p.Payload = map[string]string{"inbox_type": "approval", "inbox_title": p.Body, "workspace_id": "example-workspace"}
		return p
	}
	if event == EventSystemUpdateAvailable {
		p.Payload = map[string]string{"version": "1.2.3", "url": "https://example.invalid/releases/1.2.3"}
		p.Title, p.Body = semanticMessageCopy(event, "1.2.3")
		for _, status := range []string{updateStatusAvailable, "succeeded", "failed", "interrupted"} {
			if sample == "runtime_"+status {
				p.Payload = map[string]string{"agent_name": "example-agent", "display_name": "Example Agent", "runtime_id": "example-runtime", "previous_version": "1.0.0", "version": "1.2.3", "runtime_update_status": status}
				p.Title, p.Body = runtimeUpdateMessage(p.Payload)
			}
		}
		if sample == "runtime_summary" {
			p.RuntimeUpdates = []models.RuntimeUpdateMember{{AgentName: "example-a"}, {AgentName: "example-b"}}
			p.Title, p.Body = "2 agent runtime updates available", "Review runtime versions in Settings > Agents."
		}
		return p
	}
	p.TaskID, p.TaskSessionID = "example-task", "example-session"
	p.Title, p.Body = semanticMessageCopy(event, "Example task")
	if sample == "missing_session" {
		p.TaskSessionID = ""
	}
	if sample == "missing_title" {
		p.TaskID = "missing-title"
		p.Title, p.Body = semanticMessageCopy(event, "")
	}
	return p
}

type sampleTasks struct{}

func (sampleTasks) GetTask(_ context.Context, id string) (*taskmodels.Task, error) {
	title := "Example task"
	if id == "missing-title" {
		title = ""
	}
	return &taskmodels.Task{ID: id, Title: title, WorkspaceID: "example-workspace"}, nil
}
func (sampleTasks) GetWorkspace(context.Context, string) (*taskmodels.Workspace, error) {
	return nil, fmt.Errorf("preview never resolves recipients")
}
func (sampleTasks) GetTaskSession(context.Context, string) (*taskmodels.TaskSession, error) {
	return &taskmodels.TaskSession{ID: "example-session", TaskID: "example-task"}, nil
}
