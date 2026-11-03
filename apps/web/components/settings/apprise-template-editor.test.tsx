import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { NotificationProvider } from "@/lib/types/http";
import type {
  NotificationTemplate,
  TemplateCatalog,
} from "@/lib/api/domains/notification-templates-api";
import { SettingsSaveProvider } from "./settings-save-provider";
import { AppriseTemplateEditor } from "./apprise-template-editor";

const provider: NotificationProvider = {
  id: "example",
  name: "Example provider",
  type: "apprise",
  enabled: false,
  events: [],
  config: {},
  created_at: "2026-11-03T00:00:00Z",
  updated_at: "2026-11-03T00:00:00Z",
};
const event = "session.clarification_requested";
const saveLabel = "Save changes";
const titleLabel = "Title template";
function setup() {
  let saved: NotificationTemplate | null = null;
  let fail = false;
  let hold: (() => Promise<void>) | null = null;
  let previewGate: Promise<void> | null = null;
  const calls: string[] = [];
  const fetchMock = vi.fn(async (input: string | URL | Request, init?: RequestInit) => {
    const url = String(input);
    const method = init?.method ?? "GET";
    calls.push(`${method} ${url}`);
    if (method === "GET") {
      const catalog: TemplateCatalog = {
        public_url_configured: false,
        events: [
          {
            event,
            variables: ["task.title", "session.url", "default.title", "default.body"],
            samples: ["normal", "missing_session"],
            default_title: "Agent needs your answer",
            default_body: "The agent asked a question.",
            override: saved,
          },
        ],
      };
      return new Response(JSON.stringify(catalog), { status: 200 });
    }
    if (url.endsWith("/preview")) {
      if (previewGate) await previewGate;
      return new Response(
        JSON.stringify({
          title: "Synthetic title",
          body: "Synthetic body",
          used_fallback: false,
          diagnostics: [],
        }),
      );
    }
    if (fail) return new Response(JSON.stringify({ error: "invalid_template" }), { status: 400 });
    if (method === "DELETE") {
      saved = null;
      return new Response(null, { status: 204 });
    }
    const body = JSON.parse(String(init?.body));
    saved = { version: 1, title: body.title, body: body.body, revision: "revision-1" };
    const response = new Response(JSON.stringify(saved));
    if (hold) await hold();
    return response;
  });
  vi.stubGlobal("fetch", fetchMock);
  render(
    <SettingsSaveProvider>
      <AppriseTemplateEditor provider={provider} />
    </SettingsSaveProvider>,
  );
  return {
    calls,
    getSaved: () => saved,
    fail: () => {
      fail = true;
    },
    hold: (fn: () => Promise<void>) => {
      hold = fn;
    },
    holdPreview: (gate: Promise<void>) => {
      previewGate = gate;
    },
  };
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("Apprise template settings", () => {
  it("ignores a preview response after its draft changes", async () => {
    const api = setup();
    let release!: () => void;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    api.holdPreview(gate);
    const title = await screen.findByLabelText(titleLabel);
    fireEvent.click(screen.getByRole("button", { name: "Preview template" }));
    await waitFor(() => expect(api.calls.some((call) => call.endsWith("/preview"))).toBe(true));
    fireEvent.change(title, { target: { value: "Newer draft" } });
    await act(async () => {
      release();
      await gate;
    });
    expect(screen.queryByText("Synthetic title")).toBeNull();
    expect((title as HTMLInputElement).value).toBe("Newer draft");
  });
  it("saves overrides independently of subscriptions, previews without test send, and restores defaults", async () => {
    const api = setup();
    const title = await screen.findByLabelText(titleLabel);
    expect(screen.getByText(/Not subscribed: this event will not be sent/)).toBeTruthy();
    fireEvent.change(title, { target: { value: "Custom {{task.title}}" } });
    fireEvent.click(screen.getByRole("button", { name: "Preview template" }));
    await screen.findByText("Synthetic title");
    expect(api.getSaved()).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: saveLabel }));
    await waitFor(() => expect(api.getSaved()?.title).toBe("Custom {{task.title}}"));
    fireEvent.click(screen.getByRole("button", { name: "Restore default content" }));
    expect(api.getSaved()).not.toBeNull();
    fireEvent.click(screen.getByRole("button", { name: saveLabel }));
    await waitFor(() => expect(api.getSaved()).toBeNull());
    expect(api.calls.some((call) => call.endsWith("/test"))).toBe(false);
    expect(api.calls.filter((call) => call.startsWith("PUT"))).toHaveLength(1);
  });

  it("preserves edits and exposes a failed save", async () => {
    const api = setup();
    const title = await screen.findByLabelText(titleLabel);
    fireEvent.change(title, { target: { value: "{{unknown.variable}}" } });
    api.fail();
    fireEvent.click(screen.getByRole("button", { name: saveLabel }));
    await screen.findByText(/Could not save the template/);
    expect((title as HTMLInputElement).value).toBe("{{unknown.variable}}");
    expect(api.getSaved()).toBeNull();
  });

  it("keeps later edits after a held save acknowledgement", async () => {
    const api = setup();
    let release!: () => void;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    api.hold(() => gate);
    const title = await screen.findByLabelText(titleLabel);
    fireEvent.change(title, { target: { value: "Submitted" } });
    fireEvent.click(screen.getByRole("button", { name: saveLabel }));
    await waitFor(() => expect(api.getSaved()?.title).toBe("Submitted"));
    fireEvent.change(title, { target: { value: "Later edit" } });
    await act(async () => {
      release();
      await gate;
    });
    expect((title as HTMLInputElement).value).toBe("Later edit");
    await waitFor(() => expect(screen.getByRole("button", { name: saveLabel })).toBeTruthy());
  });
});
