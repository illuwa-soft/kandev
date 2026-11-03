import { useTranslation } from "react-i18next";
import { Button } from "@kandev/ui/button";
import { Input } from "@kandev/ui/input";
import { Textarea } from "@kandev/ui/textarea";
import type { TemplateEvent } from "@/lib/api/domains/notification-templates-api";
import type { useAppriseTemplates } from "@/hooks/domains/settings/use-apprise-templates";

type Props = {
  editor: ReturnType<typeof useAppriseTemplates>;
  entry: TemplateEvent;
  prefix: string;
};

export function AppriseTemplateFields({ editor, entry, prefix }: Props) {
  const { t } = useTranslation();
  const draft = editor.state.drafts[entry.event];
  const title = draft?.title ?? "{{default.title}}";
  const body = draft?.body ?? "{{default.body}}";
  return (
    <>
      <label className="block" htmlFor={`${prefix}-title`}>
        {t("settings:appriseTemplates.subject")}
      </label>
      <Input
        id={`${prefix}-title`}
        value={title}
        aria-describedby={editor.state.error ? `${prefix}-error` : undefined}
        onChange={(event) => editor.edit({ version: 1, title: event.target.value, body })}
      />
      <label className="block" htmlFor={`${prefix}-body`}>
        {t("settings:appriseTemplates.body")}
      </label>
      <Textarea
        id={`${prefix}-body`}
        value={body}
        rows={5}
        aria-describedby={`${prefix}-help${editor.state.error ? ` ${prefix}-error` : ""}`}
        onChange={(event) => editor.edit({ version: 1, title, body: event.target.value })}
      />
      <p id={`${prefix}-help`} className="text-sm text-muted-foreground">
        {t("settings:appriseTemplates.constraints")}
      </p>
      <p className="text-sm">{t("settings:appriseTemplates.variables")}</p>
      <div className="flex flex-wrap gap-2">
        {entry.variables.map((variable) => (
          <Button
            key={variable}
            variant="outline"
            size="sm"
            className="cursor-pointer min-h-11 md:min-h-7 [@media(pointer:coarse)]:min-h-11 max-w-full whitespace-normal break-all"
            onClick={() => editor.edit({ version: 1, title, body: `${body}{{${variable}}}` })}
          >
            {variable}
          </Button>
        ))}
      </div>
      <p className="text-sm text-muted-foreground">{t("settings:appriseTemplates.availability")}</p>
      <details>
        <summary className="cursor-pointer min-h-11 md:min-h-7 [@media(pointer:coarse)]:min-h-11">
          {t("settings:appriseTemplates.defaultExample")}
        </summary>
        <p>{entry.default_title}</p>
        <pre className="whitespace-pre-wrap break-words">{entry.default_body}</pre>
      </details>
      <Button
        variant="outline"
        className="cursor-pointer min-h-11 md:min-h-7 [@media(pointer:coarse)]:min-h-11"
        onClick={() => editor.edit(null)}
      >
        {t("settings:appriseTemplates.restore")}
      </Button>
    </>
  );
}

export function AppriseTemplatePreview({ editor, entry, prefix }: Props) {
  const { t } = useTranslation();
  const { state } = editor;
  return (
    <>
      <label className="block" htmlFor={`${prefix}-sample`}>
        {t("settings:appriseTemplates.sample")}
      </label>
      <select
        id={`${prefix}-sample`}
        className="w-full rounded border bg-background p-2 cursor-pointer min-h-11 md:min-h-7 [@media(pointer:coarse)]:min-h-11"
        value={state.sample}
        onChange={(event) => editor.selectSample(event.target.value)}
      >
        {entry.samples.map((sample) => (
          <option key={sample} value={sample}>
            {t(`settings:appriseTemplates.samples.${sample}`)}
          </option>
        ))}
      </select>
      <Button
        variant="outline"
        className="cursor-pointer min-h-11 md:min-h-7 [@media(pointer:coarse)]:min-h-11"
        disabled={state.previewPending}
        onClick={() => void editor.preview()}
      >
        {t("settings:appriseTemplates.preview")}
      </Button>
      <p role="status">
        {t(
          state.previewPending
            ? "settings:appriseTemplates.previewLoading"
            : "settings:appriseTemplates.synthetic",
        )}
      </p>
      {state.preview && (
        <div data-testid="apprise-template-preview" className="rounded border p-3 space-y-2">
          {state.preview.used_fallback && (
            <p role="status">{t("settings:appriseTemplates.fallback")}</p>
          )}
          <p className="break-words">{state.preview.title}</p>
          <pre className="whitespace-pre-wrap break-words">{state.preview.body}</pre>
        </div>
      )}
    </>
  );
}
