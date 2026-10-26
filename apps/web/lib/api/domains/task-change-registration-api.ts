import { fetchJson, type ApiRequestOptions } from "../client";

export type TaskChangeRegistrationAction = "retry" | "relink";

export function actTaskChangeRegistration(
  taskId: string,
  registrationId: string,
  action: TaskChangeRegistrationAction,
  options?: ApiRequestOptions,
) {
  return fetchJson<{ accepted: boolean; state: "pending" }>(
    `/api/v1/tasks/${encodeURIComponent(taskId)}/change-request-registrations/${encodeURIComponent(registrationId)}/retry`,
    {
      ...options,
      init: {
        ...(options?.init ?? {}),
        method: "POST",
        body: JSON.stringify({ action, operation_id: createOperationId() }),
      },
    },
  );
}

export function retryTaskChangeRegistration(
  taskId: string,
  registrationId: string,
  options?: ApiRequestOptions,
) {
  return actTaskChangeRegistration(taskId, registrationId, "retry", options);
}

function createOperationId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return `${Date.now().toString(36)}${Math.random().toString(36).slice(2)}`;
}
