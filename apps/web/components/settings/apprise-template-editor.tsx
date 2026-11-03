import { useTranslation } from "react-i18next";
import { Button } from "@kandev/ui/button";
import { useAppriseTemplates } from "@/hooks/domains/settings/use-apprise-templates";
import type { NotificationProvider } from "@/lib/types/http";
import { useSettingsSaveContributor } from "./settings-save-provider";
import { SettingsGroup } from "./settings-group";
import { AppriseTemplateFields, AppriseTemplatePreview } from "./apprise-template-controls";

export function AppriseTemplateEditor({ provider }: { provider: NotificationProvider }) {
  const { t } = useTranslation();
  const editor = useAppriseTemplates(provider.id);
  const { state } = editor;
  const entry = state.catalog?.events.find((item) => item.event === state.event);
  const draft = state.drafts[state.event];
  const prefix = `apprise-template-${provider.id}`;
  useSettingsSaveContributor({
    id: prefix,
    revision: editor.revision,
    isDirty: editor.isDirty,
    save: editor.save,
    discard: editor.discard,
  });
  return (
    <SettingsGroup
      title={t("settings:appriseTemplates.title")}
      description={t("settings:appriseTemplates.help")}
      isDirty={editor.isDirty}
      data-testid="apprise-template-editor"
    >
      <div className="space-y-3 min-w-0">
        <p className="font-medium break-words">{provider.name}</p>
        {state.error && (
          <p role="alert" id={`${prefix}-error`}>
            {t(`settings:appriseTemplates.${state.error}Error`)}
          </p>
        )}
        {(state.error === "load" || state.error === "conflict") && (
          <Button
            variant="outline"
            className="cursor-pointer min-h-11 md:min-h-7"
            onClick={editor.reload}
          >
            {t("settings:appriseTemplates.reload")}
          </Button>
        )}
        {!state.catalog && !state.error && (
          <p role="status">{t("settings:appriseTemplates.loading")}</p>
        )}
        {entry && (
          <>
            <label className="block" htmlFor={`${prefix}-event`}>
              {t("settings:appriseTemplates.event")}
            </label>
            <select
              id={`${prefix}-event`}
              className="w-full rounded border bg-background p-2 cursor-pointer min-h-11 md:min-h-7"
              value={state.event}
              onChange={(event) => editor.selectEvent(event.target.value)}
            >
              {state.catalog?.events.map((item) => (
                <option key={item.event} value={item.event}>
                  {item.event}
                </option>
              ))}
            </select>
            <p>
              {t(
                provider.enabled && provider.events?.includes(state.event)
                  ? "settings:appriseTemplates.subscribed"
                  : "settings:appriseTemplates.unsubscribed",
              )}{" "}
              ·{" "}
              {t(draft ? "settings:appriseTemplates.custom" : "settings:appriseTemplates.default")}
            </p>
            {!state.catalog?.public_url_configured && (
              <p role="status">{t("settings:appriseTemplates.noUrl")}</p>
            )}
            <AppriseTemplateFields editor={editor} entry={entry} prefix={prefix} />
            <AppriseTemplatePreview editor={editor} entry={entry} prefix={prefix} />
          </>
        )}
      </div>
    </SettingsGroup>
  );
}
