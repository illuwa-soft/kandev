import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
  type RefObject,
} from "react";
import { ApiError } from "@/lib/api/client";
import {
  loadTemplateCatalog,
  previewNotificationTemplate,
  saveNotificationTemplate,
  type TemplateCatalog,
  type TemplateDraft,
  type TemplatePreview,
} from "@/lib/api/domains/notification-templates-api";

type EditorState = {
  catalog: TemplateCatalog | null;
  drafts: Record<string, TemplateDraft>;
  event: string;
  sample: string;
  preview: TemplatePreview | null;
  previewPending: boolean;
  error: "load" | "save" | "preview" | "conflict" | null;
};

function initialState(): EditorState {
  return {
    catalog: null,
    drafts: {},
    event: "",
    sample: "normal",
    preview: null,
    previewPending: false,
    error: null,
  };
}
function draftValue(value: TemplateDraft) {
  return value ? { version: value.version, title: value.title, body: value.body } : null;
}
function sameDraft(a: TemplateDraft, b: TemplateDraft) {
  return JSON.stringify(draftValue(a)) === JSON.stringify(draftValue(b));
}
function dirtyEvents(state: EditorState) {
  return (
    state.catalog?.events.filter(
      (entry) => !sameDraft(state.drafts[entry.event], entry.override),
    ) ?? []
  );
}

type SetEditor = Dispatch<SetStateAction<EditorState>>;
function useTemplateLoad(
  providerId: string,
  reloadVersion: number,
  generation: RefObject<number>,
  setState: SetEditor,
) {
  useEffect(() => {
    const request = ++generation.current;
    setState(initialState());
    void loadTemplateCatalog(providerId)
      .then((catalog) => {
        if (request !== generation.current) return;
        setState({
          ...initialState(),
          catalog,
          event: catalog.events[0]?.event ?? "",
          drafts: Object.fromEntries(
            catalog.events.map((entry) => [entry.event, draftValue(entry.override)]),
          ),
        });
      })
      .catch(() => {
        if (request === generation.current) setState((value) => ({ ...value, error: "load" }));
      });
    return () => {
      generation.current++;
    };
  }, [providerId, reloadVersion, generation, setState]);
}

async function saveTemplates(providerId: string, submitted: EditorState, setState: SetEditor) {
  setState((value) => ({ ...value, error: null }));
  try {
    for (const entry of dirtyEvents(submitted)) {
      const draft = submitted.drafts[entry.event];
      const saved = await saveNotificationTemplate(
        providerId,
        entry.event,
        draft,
        entry.override?.revision ?? null,
      );
      setState((value) => ({
        ...value,
        catalog: value.catalog
          ? {
              ...value.catalog,
              events: value.catalog.events.map((item) =>
                item.event === entry.event ? { ...item, override: saved ?? null } : item,
              ),
            }
          : null,
        drafts: sameDraft(value.drafts[entry.event], draft)
          ? { ...value.drafts, [entry.event]: draftValue(saved ?? null) }
          : value.drafts,
      }));
    }
  } catch (error) {
    setState((value) => ({
      ...value,
      error: error instanceof ApiError && error.status === 409 ? "conflict" : "save",
    }));
    throw error;
  }
}

export function useAppriseTemplates(providerId: string) {
  const [state, setState] = useState<EditorState>(initialState);
  const current = useRef(state);
  current.current = state;
  const generation = useRef(0);
  const [reloadVersion, setReloadVersion] = useState(0);
  useTemplateLoad(providerId, reloadVersion, generation, setState);

  const edit = useCallback((draft: TemplateDraft) => {
    generation.current++;
    setState((value) => ({
      ...value,
      drafts: { ...value.drafts, [value.event]: draft },
      preview: null,
      previewPending: false,
      error: null,
    }));
  }, []);
  const selectEvent = useCallback((event: string) => {
    generation.current++;
    setState((value) => ({
      ...value,
      event,
      sample: "normal",
      preview: null,
      previewPending: false,
      error: null,
    }));
  }, []);
  const selectSample = useCallback((sample: string) => {
    generation.current++;
    setState((value) => ({ ...value, sample, preview: null, previewPending: false }));
  }, []);

  const save = useCallback(
    () => saveTemplates(providerId, current.current, setState),
    [providerId],
  );

  const preview = useCallback(async () => {
    const value = current.current;
    const entry = value.catalog?.events.find((item) => item.event === value.event);
    if (!entry) return;
    const draft = value.drafts[value.event] ?? {
      version: 1,
      title: "{{default.title}}",
      body: "{{default.body}}",
    };
    const request = ++generation.current;
    setState((previous) => ({ ...previous, previewPending: true, error: null }));
    try {
      const result = await previewNotificationTemplate(
        providerId,
        value.event,
        draft,
        value.sample,
      );
      if (request === generation.current)
        setState((previous) => ({ ...previous, preview: result }));
    } catch {
      if (request === generation.current)
        setState((previous) => ({ ...previous, error: "preview" }));
    } finally {
      if (request === generation.current)
        setState((previous) => ({ ...previous, previewPending: false }));
    }
  }, [providerId]);

  const discard = useCallback(() => {
    generation.current++;
    setState((value) => ({
      ...value,
      drafts: Object.fromEntries(
        value.catalog?.events.map((entry) => [entry.event, draftValue(entry.override)]) ?? [],
      ),
      preview: null,
      previewPending: false,
      error: null,
    }));
  }, []);
  return {
    state,
    edit,
    selectEvent,
    selectSample,
    save,
    preview,
    discard,
    reload: () => setReloadVersion((value) => value + 1),
    isDirty: dirtyEvents(state).length > 0,
    revision: JSON.stringify(state.drafts),
  };
}
