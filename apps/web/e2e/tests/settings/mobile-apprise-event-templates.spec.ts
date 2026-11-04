import { test } from "../../fixtures/test-base";
import { appriseTemplateScenario } from "./apprise-event-templates-helpers";

test("manages Apprise templates on a phone without horizontal scrolling", async ({
  testPage,
  apiClient,
}, info) => {
  await appriseTemplateScenario(testPage, apiClient, info);
});
