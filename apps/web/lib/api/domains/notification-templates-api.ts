import { fetchJson } from "../client";

export type NotificationTemplate = {
  version: number;
  title: string;
  body: string;
  revision: string;
};
export type TemplateEvent = {
  event: string;
  variables: string[];
  samples: string[];
  default_title: string;
  default_body: string;
  override: NotificationTemplate | null;
};
export type TemplateCatalog = { events: TemplateEvent[]; public_url_configured: boolean };
export type TemplatePreview = {
  title: string;
  body: string;
  used_fallback: boolean;
  diagnostics: string[];
};
export type TemplateDraft = { version: number; title: string; body: string } | null;

function templatePath(providerId: string, event?: string) {
  const base = `/api/v1/notification-providers/${encodeURIComponent(providerId)}/templates`;
  return event ? `${base}/${encodeURIComponent(event)}` : base;
}

export function loadTemplateCatalog(providerId: string) {
  return fetchJson<TemplateCatalog>(templatePath(providerId), { cache: "no-store" });
}
export function saveNotificationTemplate(
  providerId: string,
  event: string,
  draft: TemplateDraft,
  revision: string | null,
) {
  return fetchJson<NotificationTemplate | undefined>(templatePath(providerId, event), {
    init: {
      method: draft ? "PUT" : "DELETE",
      body: JSON.stringify({ ...draft, expected_revision: revision }),
    },
  });
}
export function previewNotificationTemplate(
  providerId: string,
  event: string,
  draft: NonNullable<TemplateDraft>,
  sample: string,
) {
  return fetchJson<TemplatePreview>(`${templatePath(providerId, event)}/preview`, {
    init: { method: "POST", body: JSON.stringify({ ...draft, sample }) },
  });
}
