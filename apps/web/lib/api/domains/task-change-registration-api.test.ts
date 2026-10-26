import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  actTaskChangeRegistration,
  retryTaskChangeRegistration,
} from "./task-change-registration-api";

const fetchSpy = vi.fn<typeof fetch>();

beforeEach(() => {
  fetchSpy.mockReset();
  vi.stubGlobal("fetch", fetchSpy);
});

afterEach(() => vi.unstubAllGlobals());

describe("task change registration actions", () => {
  it.each(["retry", "relink"] as const)(
    "submits the %s action without resending the message",
    async (action) => {
      fetchSpy.mockResolvedValueOnce(
        new Response(JSON.stringify({ accepted: true, state: "pending" }), {
          status: 202,
          headers: { "Content-Type": "application/json" },
        }),
      );

      const request =
        action === "retry"
          ? retryTaskChangeRegistration("task/1", "selection-1", { baseUrl: "http://api.test" })
          : actTaskChangeRegistration("task/1", "selection-1", action, {
              baseUrl: "http://api.test",
            });
      await expect(request).resolves.toEqual({ accepted: true, state: "pending" });
      expect(fetchSpy.mock.calls[0][0]).toBe(
        "http://api.test/api/v1/tasks/task%2F1/change-request-registrations/selection-1/retry",
      );
      const init = fetchSpy.mock.calls[0][1];
      expect(init?.method).toBe("POST");
      const body = JSON.parse(String(init?.body)) as { action: string; operation_id: string };
      expect(body.action).toBe(action);
      expect(body.operation_id).toBeTruthy();
    },
  );
});
