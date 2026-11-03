import { expect, type Page, type TestInfo } from "@playwright/test";
import type { ApiClient } from "../../helpers/api-client";
import { waitForHttp } from "../../helpers/causal-waits";

export async function appriseTemplateScenario(page: Page, api: ApiClient, info: TestInfo) {
  const response = await api.rawRequest("POST", "/api/v1/notification-providers", {
    name: "E2E template provider",
    type: "apprise",
    config: { urls: ["json://synthetic-receiver.invalid"] },
    enabled: false,
    events: [],
  });
  expect(response.ok).toBe(true);
  const { id } = await response.json();
  try {
    await page.goto("/settings/preferences/notifications");
    const editor = page
      .getByTestId("apprise-template-editor")
      .filter({ hasText: "E2E template provider" });
    const title = editor.getByLabel("Title template", { exact: true });
    const body = editor.getByLabel("Body template", { exact: true });
    await expect(title).toBeVisible();
    await editor
      .getByLabel("Notification event", { exact: true })
      .selectOption("session.clarification_requested");
    await expect(editor.getByText("Not subscribed: this event will not be sent")).toBeVisible();
    await title.fill("Answer: {{task.title}}");
    await body.fill("{{default.body}}");
    const previewResponse = waitForHttp(
      page,
      "POST",
      /\/templates\/session.clarification_requested\/preview$/,
    );
    await editor.getByRole("button", { name: "Preview template", exact: true }).click();
    await previewResponse;
    await expect(editor.getByTestId("apprise-template-preview")).toContainText(
      "Answer: Example task",
    );
    await info.attach("synthetic-template-preview", {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    const save = page.getByRole("button", { name: /^(Save changes|Retry save)$/ });
    const saved = waitForHttp(page, "PUT", /\/templates\/session.clarification_requested$/);
    await save.click();
    await saved;
    await page.reload();
    await editor
      .getByLabel("Notification event", { exact: true })
      .selectOption("session.clarification_requested");
    await expect(title).toHaveValue("Answer: {{task.title}}");
    await title.fill("{{unknown.variable}}");
    const invalid = waitForHttp(page, "PUT", /\/templates\/session.clarification_requested$/);
    await save.click();
    expect((await invalid).status()).toBe(400);
    await expect(editor.getByRole("alert")).toContainText("Could not save");
    await expect(title).toHaveValue("{{unknown.variable}}");
    await editor.getByRole("button", { name: "Restore default content", exact: true }).click();
    const deleted = waitForHttp(page, "DELETE", /\/templates\/session.clarification_requested$/);
    await save.click();
    await deleted;
    await page.reload();
    await editor
      .getByLabel("Notification event", { exact: true })
      .selectOption("session.clarification_requested");
    await expect(title).toHaveValue("{{default.title}}");
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
    ).toBe(true);
  } finally {
    expect((await api.rawRequest("DELETE", `/api/v1/notification-providers/${id}`)).ok).toBe(true);
  }
}
