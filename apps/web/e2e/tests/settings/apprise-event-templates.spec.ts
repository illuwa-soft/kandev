import { test } from "../../fixtures/test-base";
import { appriseTemplateScenario } from "./apprise-event-templates-helpers";

test("manages Apprise event templates without sending notifications", async ({
  testPage,
  apiClient,
}, info) => {
  await appriseTemplateScenario(testPage, apiClient, info);
});
