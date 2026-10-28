import { useEffect, useRef } from "react";
import type { TurnFileChange } from "@/lib/types/turn-changes";

export function useReadyContentRetry(
  selectedFile: TurnFileChange | null,
  contentError: unknown,
  contentLoading: boolean,
  whitespace: boolean,
  retry: () => void,
) {
  const retried = useRef(new Set<string>());
  useEffect(() => {
    if (
      !selectedFile ||
      selectedFile.content_availability !== "ready" ||
      !contentError ||
      contentLoading
    )
      return;
    const key = `${selectedFile.id}:${whitespace}`;
    if (retried.current.has(key)) return;
    retried.current.add(key);
    retry();
  }, [contentError, contentLoading, retry, selectedFile, whitespace]);
}
