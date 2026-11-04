package service

import (
	"context"
	"testing"

	"github.com/kandev/kandev/internal/notifications/models"
	"github.com/stretchr/testify/require"
)

func TestEveryEventCustomDefaultFallbackAndSubscription(t *testing.T) {
	t.Setenv("KANDEV_DESKTOP_NATIVE_NOTIFICATIONS", "true")
	for _, event := range []string{EventTaskSessionTurnFinished, EventTaskSessionClarificationAsked, EventOfficeInboxItem, EventSystemUpdateAvailable} {
		t.Run(event, func(t *testing.T) {
			svc, repo, database := atomicityService(t)
			ctx := context.Background()
			svc.taskRepo = sampleTasks{}
			require.NoError(t, svc.SetPublicBaseURL("https://example.invalid"))
			capture := &captureProvider{}
			svc.providers[models.ProviderTypeApprise] = capture
			provider, err := svc.CreateProvider(ctx, "owner", "test", models.ProviderTypeApprise, nil, true, []string{event})
			require.NoError(t, err)
			original := samplePayload(event, "normal")
			original.OccurrenceID = "default"
			svc.deliverOccurrence(ctx, "owner", original)
			require.Len(t, capture.messages, 1)
			require.Equal(t, original.Title, capture.messages[0].Title)
			require.Equal(t, original.Body, capture.messages[0].Body)
			saved, err := svc.SaveTemplate(ctx, "owner", provider.ID, event, "", &models.NotificationTemplate{Version: 1, Title: "Custom {{event.type}}", Body: "{{default.body}} {{event.url}}"})
			require.NoError(t, err)
			original.OccurrenceID = "custom"
			svc.deliverOccurrence(ctx, "owner", original)
			require.Len(t, capture.messages, 2)
			require.Equal(t, "Custom "+event, capture.messages[1].Title)
			require.Contains(t, capture.messages[1].Body, "https://example.invalid/")
			require.NoError(t, svc.SetPublicBaseURL(""))
			original.OccurrenceID = "fallback"
			svc.deliverOccurrence(ctx, "owner", original)
			require.Equal(t, original.Title, capture.messages[2].Title)
			require.Equal(t, original.Body, capture.messages[2].Body)
			require.NoError(t, repo.ReplaceSubscriptions(ctx, provider.ID, "owner", []string{}))
			original.OccurrenceID = "unsubscribed"
			svc.deliverOccurrence(ctx, "owner", original)
			require.Len(t, capture.messages, 3)
			require.NoError(t, repo.ReplaceSubscriptions(ctx, provider.ID, "owner", []string{event}))
			_, err = database.Exec(`UPDATE notification_templates SET template_json = 'invalid' WHERE provider_id = ?`, provider.ID)
			require.NoError(t, err)
			original.OccurrenceID = "invalid-storage"
			svc.deliverOccurrence(ctx, "owner", original)
			require.Equal(t, original.Body, capture.messages[3].Body)
			require.NoError(t, svc.DeleteTemplate(ctx, "owner", provider.ID, event, saved.Revision))
		})
	}
}

func TestSessionVariablesNeverBorrowOtherTaskSession(t *testing.T) {
	svc, _, _ := atomicityService(t)
	svc.taskRepo = sampleTasks{}
	require.NoError(t, svc.SetPublicBaseURL("https://example.invalid"))
	for _, sample := range []string{"normal", "missing_session", "missing_title"} {
		p := samplePayload(EventTaskSessionClarificationAsked, sample)
		values := svc.templateValues(context.Background(), p)
		if sample != "normal" {
			require.Empty(t, values["session.id"])
			require.Empty(t, values["session.url"])
		} else {
			require.Equal(t, "example-session", values["session.id"])
			require.Equal(t, "https://example.invalid/t/example-task?sessionId=example-session", values["session.url"])
		}
	}
}

func TestRuntimeTemplateShapesAndSyntheticPreview(t *testing.T) {
	svc, repo, database := atomicityService(t)
	capture := &captureProvider{}
	svc.providers[models.ProviderTypeApprise] = capture
	ctx := context.Background()
	p, err := svc.CreateProvider(ctx, "owner", "test", models.ProviderTypeApprise, nil, false, nil)
	require.NoError(t, err)
	require.NoError(t, svc.SetPublicBaseURL("https://example.invalid"))
	for _, sample := range templateSamples(EventSystemUpdateAvailable) {
		preview, err := svc.PreviewTemplate(ctx, "owner", p.ID, EventSystemUpdateAvailable, sample, &models.NotificationTemplate{Version: 1, Title: "{{default.title}}", Body: "{{update.kind}} {{event.url}}"})
		require.NoError(t, err)
		require.Equal(t, sample == "missing_url", preview.UsedFallback)
		if sample == "runtime_summary" {
			require.Contains(t, preview.Body, "agent_runtime_summary")
			require.Contains(t, preview.Body, "#runtime-updates")
		}
	}
	for _, event := range svc.AvailableEvents() {
		_, err := svc.PreviewTemplate(ctx, "owner", p.ID, event, "normal", &models.NotificationTemplate{Version: 1, Title: "{{default.title}}", Body: "{{default.body}}"})
		require.NoError(t, err)
	}
	_, err = svc.PreviewTemplate(ctx, "owner", p.ID, EventSystemUpdateAvailable, "normal", &models.NotificationTemplate{Version: 1, Title: "{{session.id}}", Body: "body"})
	require.Error(t, err)
	_, err = svc.PreviewTemplate(ctx, "foreign", p.ID, EventSystemUpdateAvailable, "normal", &models.NotificationTemplate{Version: 1, Title: "title", Body: "body"})
	require.Error(t, err)
	require.Empty(t, capture.messages)
	var count int
	require.NoError(t, database.Get(&count, "SELECT COUNT(*) FROM notification_deliveries"))
	require.Zero(t, count)
	stored, err := repo.GetTemplate(ctx, "owner", p.ID, EventSystemUpdateAvailable)
	require.NoError(t, err)
	require.Nil(t, stored)
}

func TestOtherProvidersAndFixedTestSendKeepDefaultCopy(t *testing.T) {
	t.Setenv("KANDEV_DESKTOP_NATIVE_NOTIFICATIONS", "true")
	svc, repo, _ := atomicityService(t)
	ctx := context.Background()
	external, local := &captureProvider{}, &captureProvider{}
	svc.providers[models.ProviderTypeApprise] = external
	svc.providers[models.ProviderTypeLocal] = local
	p, err := svc.CreateProvider(ctx, "owner", "external", models.ProviderTypeApprise, nil, true, []string{EventTaskSessionTurnFinished})
	require.NoError(t, err)
	_, err = svc.CreateProvider(ctx, "owner", "local", models.ProviderTypeLocal, nil, true, []string{EventTaskSessionTurnFinished})
	require.NoError(t, err)
	_, err = svc.SaveTemplate(ctx, "owner", p.ID, EventTaskSessionTurnFinished, "", &models.NotificationTemplate{Version: 1, Title: "custom", Body: "custom body"})
	require.NoError(t, err)
	message := samplePayload(EventTaskSessionTurnFinished, "normal")
	message.OccurrenceID = "control"
	svc.deliverOccurrence(ctx, "owner", message)
	require.Equal(t, "custom", external.messages[0].Title)
	require.Equal(t, message.Title, local.messages[0].Title)
	require.Equal(t, message.Body, local.messages[0].Body)
	p.Enabled = false
	require.NoError(t, repo.UpdateProviderWithSubscriptions(ctx, p, nil))
	message.OccurrenceID = "disabled"
	svc.deliverOccurrence(ctx, "owner", message)
	require.Len(t, external.messages, 1)
	require.NoError(t, svc.TestProvider(ctx, "owner", p.ID))
	require.Equal(t, "Test notification", external.messages[1].Title)
}
