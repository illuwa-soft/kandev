package service

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kandev/kandev/internal/notifications/templates"
	taskmodels "github.com/kandev/kandev/internal/task/models"
)

type templateSessionReader interface {
	GetTaskSession(context.Context, string) (*taskmodels.TaskSession, error)
}

func (s *Service) templateValues(ctx context.Context, p notificationPayload) map[string]string {
	origin := s.publicOrigin()
	v := map[string]string{"event.type": p.EventType, "default.title": p.Title, "default.body": p.Body}
	switch p.EventType {
	case EventTaskSessionTurnFinished, EventTaskSessionClarificationAsked:
		s.taskTemplateValues(ctx, p, v, origin)
	case EventOfficeInboxItem:
		v["inbox.title"], v["inbox.type"] = p.Payload["inbox_title"], p.Payload["inbox_type"]
		path := "/office/inbox"
		if p.Payload["workspace_id"] != "" {
			path += "?" + url.Values{"workspaceId": {p.Payload["workspace_id"]}}.Encode()
		}
		v["event.url"] = templates.Destination(origin, path)
	case EventSystemUpdateAvailable:
		updateTemplateValues(v, p, origin)
	}
	return v
}

func (s *Service) taskTemplateValues(ctx context.Context, p notificationPayload, v map[string]string, origin string) {
	if s.taskRepo == nil || p.TaskID == "" {
		return
	}
	task, err := s.taskRepo.GetTask(ctx, p.TaskID)
	if err != nil || task == nil || task.ID != p.TaskID {
		return
	}
	v["task.id"], v["task.title"] = task.ID, task.Title
	v["task.url"] = templates.TaskURL(origin, task.ID)
	if reader, ok := s.taskRepo.(templateSessionReader); ok && p.TaskSessionID != "" {
		session, err := reader.GetTaskSession(ctx, p.TaskSessionID)
		if err == nil && session != nil && session.ID == p.TaskSessionID && session.TaskID == task.ID {
			v["session.id"] = session.ID
			v["session.url"] = templates.SessionURL(origin, task.ID, session.ID)
			v["event.url"] = v["session.url"]
		}
	}
}

func updateTemplateValues(v map[string]string, p notificationPayload, origin string) {
	if len(p.RuntimeUpdates) > 1 {
		v["update.kind"], v["update.status"] = "agent_runtime_summary", "available"
		v["update.count"] = strconv.Itoa(len(p.RuntimeUpdates))
		v["event.url"] = templates.Destination(origin, "/settings/agents#runtime-updates")
		return
	}
	v["update.version"] = p.Payload["version"]
	if p.Payload["agent_name"] == "" {
		v["update.kind"] = "kandev"
		v["release.url"] = templates.ReleaseURL(p.Payload["url"])
		v["event.url"] = templates.Destination(origin, "/settings/system/updates")
		return
	}
	v["update.kind"], v["update.count"] = "agent_runtime", "1"
	v["update.agent_name"], v["update.runtime_id"] = p.Payload["display_name"], p.Payload["runtime_id"]
	v["update.previous_version"], v["update.status"] = p.Payload["previous_version"], p.Payload["runtime_update_status"]
	v["event.url"] = templates.Destination(origin, "/settings/agents#runtime-update-"+url.PathEscape(p.Payload["agent_name"]))
}
