import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { TurnFileChange } from "@/lib/types/turn-changes";
import { useReadyContentRetry } from "./use-ready-content-retry";

const notStoredError = new Error("not stored");

function file(availability: TurnFileChange["content_availability"]): TurnFileChange {
  return {
    id: "file-1",
    repository_change_id: "repo-1",
    checkout_id: "checkout-1",
    path: "src/main.ts",
    kind: "modified",
    content_availability: availability,
  };
}

describe("useReadyContentRetry", () => {
  it("does not spend the readiness retry on a synthetic pending placeholder", () => {
    const retry = vi.fn();
    const { rerender } = renderHook(
      ({ selected, whitespace, error, loading }) =>
        useReadyContentRetry(selected, error, loading, whitespace, retry),
      {
        initialProps: {
          selected: file("pending"),
          whitespace: false,
          error: notStoredError as unknown,
          loading: false,
        },
      },
    );
    expect(retry).not.toHaveBeenCalled();

    rerender({ selected: file("ready"), whitespace: false, error: notStoredError, loading: false });
    expect(retry).toHaveBeenCalledTimes(1);
    rerender({ selected: file("ready"), whitespace: false, error: notStoredError, loading: false });
    expect(retry).toHaveBeenCalledTimes(1);
    rerender({ selected: file("ready"), whitespace: true, error: notStoredError, loading: false });
    expect(retry).toHaveBeenCalledTimes(2);
  });

  it("ignores loading, unavailable, and error-free states", () => {
    const retry = vi.fn();
    const { rerender } = renderHook(
      ({ selected, whitespace, error, loading }) =>
        useReadyContentRetry(selected, error, loading, whitespace, retry),
      {
        initialProps: {
          selected: file("ready"),
          whitespace: false,
          error: new Error("loading") as unknown,
          loading: true,
        },
      },
    );
    expect(retry).not.toHaveBeenCalled();
    rerender({ selected: file("ready"), whitespace: false, error: null, loading: false });
    rerender({
      selected: file("unavailable"),
      whitespace: false,
      error: new Error("gone"),
      loading: false,
    });
    expect(retry).not.toHaveBeenCalled();
  });
});
