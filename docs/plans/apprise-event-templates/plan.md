---
created: 2026-11-03
status: in_progress
requirements:
  - REQ-PLATFORM-APPRISE-TEMPLATES-001
  - REQ-PLATFORM-APPRISE-TEMPLATES-002
  - REQ-PLATFORM-APPRISE-TEMPLATES-003
  - REQ-PLATFORM-APPRISE-TEMPLATES-004
system_design:
  - ../../specs/platform/system-design/apprise-event-templates.md
legacy_specs: []
---

# Implementation Plan: Apprise Event Templates

## Overview

Delivery baseline: Illuwa `production/mac`,
`2d8aee5817645b4e8518e56d5d268f0addac1c79`, verified against the Mac's installed
binary and deployment receipt. The branch preserves the source without changing
main. Work is isolated on `feat/apprise-event-templates`. The initial research
used newer upstream main `c060076d6723a611ebc260371aa80b555442e589`; the user
then explicitly requested implementation, a fork PR and deployment of this
feature only on the existing runtime. See [provenance and rollout](deployment.md).
The fork issue tracker is disabled; the owner's explicit fork-local direction is
recorded in the design. Upstream PR submission still requires upstream discussion.

Confirmed: all four events offered on this runtime, independent subscriptions, persistent per-owner CRUD,
bounded syntax, optional public URL, synthetic preview, defaults on deletion or
render failure, tests, public docs and eventual fork PR. Platform owns the
durable outcome. Apprise-only scope is selected under the user's instruction to
avoid unnecessary expansion. No material user question blocks drafting; API and
storage choices remain proposals pending upstream discussion.

## Scope

### In scope

- One override pair per existing Apprise provider/event, no subscription changes.
- Safe catalog, renderer, event-context destinations and startup URL configuration.
- Owned persistent APIs and settings editing using existing save coordination.
- Migration/restore, default-copy regression, security, desktop/phone E2E and docs.

### Out of scope

- Deployment, production settings or production credentials/data.
- Other providers' templates, transport replacement, new event keys and real-data preview.

## Technical approach

1. Implement the pure bounded renderer and public-origin validation with catalog
   tests. This is independently testable without sending a notification.
2. Deliver persisted CRUD/preview and event integration together. Add the new
   table to the existing store's descriptor/conformance/upgrade fixtures. Route
   inbox and all update shapes through the Apprise-only override seam. Sanitize
   Apprise subprocess diagnostics. Preserve all current claims/recipients.
3. Deliver the settings editor and API/store/hooks with shared Save/Reset, all
   translations, public setup/API/backup docs and real-backend desktop/phone E2E.

| Provider/path | Identity and intended behavior | Verification | Unsupported/failure behavior |
| --- | --- | --- | --- |
| Apprise session events | Owned provider plus verified event task/session pair; override or existing default | Per-event service matrix and fake CLI capture | Missing referenced value/render error uses default pair |
| Apprise inbox | Workspace owner, no session, inbox destination | Inbox dispatch matrix | No invented item/task/session |
| Apprise updates | Each provider owner; release/runtime/summary context | Status/subtype and claim tests | Missing subtype value uses default pair |
| Local/System/Web Push/Webhook | Current message/transport/identity unchanged | Cross-provider control assertions | Never consume Apprise templates |
| Preview | Provider owner and synthetic sample | HTTP tests, zero-send/write spy, E2E | Invalid sample rejected; no real-data fallback |

## ASCII UI preview

The following is a static specification sketch, not a rendered UI. Labels are
illustrative; structural requirements are grouping, control order, draft/save
semantics, independent subscription status and readable errors.

### UI-01: Notifications settings, shared desktop/phone composition

Entry: `/settings/preferences/notifications`, existing Apprise provider selected.

```text
Notification events
  [existing independent provider/event checkboxes]

Apprise message templates
  Provider [My Apprise v]   Event [Agent needs an answer v]
  Subscription: Off        Content: Custom
  Templates do not enable subscriptions.
  Title [Answer needed: {{task.title}}                  ]
  Body  [{{default.body}}                              ]
        [{{session.url}}                               ]
  Variables: [task.title] [session.url] [event.type] ...
  Availability: session.url requires public URL + exact session.
  [Restore defaults]    [Default example]
  Sample [Normal / missing session / missing public URL v]
  [Preview]
  Synthetic preview. No notification is sent.
  Title: Answer needed: Example task
  Body:  The agent asked a question on "Example task".
         https://example.invalid/t/example?sessionId=sample

  [shared Reset] [shared Save changes]
```

Same order on phone; provider/event controls stack, fields fill available width,
variables wrap, and action targets are at least 44px for coarse pointers. The
page owns scrolling; no desktop table is squeezed into the editor.

### UI-02: Empty, loading, fallback and failed Save

```text
No saved Apprise provider: Create and save a provider to edit templates.
Loading catalog: [announced status; no editable stale fields]
Unknown variable: Title [{{session.unknown}}] [field error]
Missing sample value: Preview uses default copy. session.url unavailable.
Restore staged: Content: Default (pending Save); fields show default example.
Save failed: Could not save. Your edits are retained. [Retry via shared Save]
Conflict: Saved template changed. Reload before retrying.
```

Maps to `AC-PLATFORM-APPRISE-TEMPLATES-003.1` through `.4`. Shared Reset discards
the staged restore, while Save applies deletion. Preview is disabled during
invalid input/loading, but does not require a subscription or installed CLI.

## Tests

All test names below are planned, not created or run in this design turn.

| Acceptance criteria | Planned evidence |
| --- | --- |
| 001.1, 001.3; 004.1 | `store/templates_test.go`: `TestTemplateCRUDIsolation`, `TestTemplateRevisionConflict` (including delete/recreate ABA), `TestTemplateRestartAndMigrationReplay`, `TestTemplateDeleteAndTypeChangeCleanup`; HTTP ownership matrix for list/PUT/DELETE/preview |
| 001.2, 001.4 | `service/apprise_templates_test.go`: `TestAppriseTemplatesRespectSubscriptions`, `TestOtherProvidersKeepDefaults`, `TestProviderTestRetainsFixedCopy` |
| 002.1, 002.5 | `templates/renderer_test.go`: `TestRendererGrammar`, `TestRendererLimits`, `TestSubstitutionsAreLiteral`; malformed-input fuzz seeds, boundary/overflow/UTF-8/title tests |
| 002.2, 002.4; 001.4 | `service/apprise_templates_test.go`: `TestEveryEventDefaultAndCustomCopy`, `TestSessionPairUnavailableNeverSubstitutesSibling`, `TestRuntimeUpdateTemplateShapes`, inbox type/title/no-session cases |
| 002.3, 002.4 | `templates/urls_test.go`: `TestPublicOriginPolicy`, `TestDestinationEncoding`, `TestAbsentOrigin`, release/runtime/inbox destinations; config catalog/source tests and link route parity |
| 002.2, 002.5 | `service/apprise_templates_test.go`: `TestRenderFailureFallsBackAsPair`; malformed saved JSON/version, repository failure and missing/oversized values |
| 003.2; 004.2 | `handlers/templates_test.go`: `TestPreviewIsSyntheticAndSideEffectFree`; `providers/apprise_test.go`: `TestAppriseErrorsDoNotExposeReceivers`, timeout/cancel/argument boundaries |
| 003.1, 003.3 | `components/settings/apprise-template-editor.test.tsx`: shared Save/Reset/restore, rejected saves, held acknowledgement/later draft and stale preview tests with real hooks/store/API |
| 004.1 | Store conformance + explicit-tag upgrade and full DB backup/restore test retaining JSON/revision; env-gated real PostgreSQL migration/replay/CRUD/race cases |
| 004.3 | Public-doc validator plus documented runtime/API examples checked against implementation |

Acceptance IDs in this table use suffixes of
`AC-PLATFORM-APPRISE-TEMPLATES-`; work orders carry full identifiers.
Use newly named test files to avoid growing large existing suites. Existing
default tests remain controls. Fake senders/CLI capture stdout internally in
disposable fixtures; no receiver credentials or live notifications.

## E2E tests

- Proposed `apps/web/e2e/tests/settings/apprise-event-templates.spec.ts`, project
  `chromium`: create override, edit, shared Save, reload, restore/delete, reload;
  preview all event shapes; invalid variable; missing origin; save failure;
  in-flight draft preservation. Covers 001.1-.2, 003.1-.3.
- Proposed `mobile-apprise-event-templates.spec.ts`, project `mobile-chrome`: same core
  workflow through native touch/stacked controls, 44px targets, no horizontal
  scrolling, field/status accessibility. Covers 003.4 and the same persistence flow.
- Use `waitForHttp`/causal waits and current `test-base` fixtures. Seed disabled
  providers with empty subscriptions; preview must produce zero transport calls.
  Use disposable real backend state and never an operator instance. Store/API
  integration tests cover backend restarts; page reload alone is not restart proof.
- Include manual desktop/phone verification and screenshots, with only synthetic
  data visible, before an eventual PR. Public docs must describe this as a feature
  only after implementation, not claim the draft setting already exists.

## Work orders

- [ ] [Task 01: Bounded catalog, renderer and URL contract](task-01-renderer-and-urls.md)
- [ ] [Task 02: Owned persistent APIs and Apprise delivery](task-02-persistence-and-delivery.md)
- [ ] [Task 03: Settings editor, documentation and E2E](task-03-settings-and-e2e.md)

All three execute sequentially; no subagents or persistent worker tasks are authorized.

## Verification results

Implementation validation on the compatibility base:

- `make fmt` and `make typecheck`: passed using Node 24.15.0 and an 8 GiB Node heap.
- `pnpm run typecheck`, scoped ESLint with `--max-warnings 0`: passed.
- Real transport/coordinator template and existing draft suites: 19 tests passed;
  the earlier combined run including external-MCP catalog controls passed 27.
- Node 24 `pnpm run i18n:check`: passed; all required locales complete.
- Go 1.26.0 `go test -trimpath -tags fts5 -race` on notifications, config and
  weburl: passed. Notification CRUD covers owner isolation, CAS/delete-recreate,
  rollback, schema replay, actual file reopen and backup restoration. The live
  PostgreSQL case is env-gated and skipped without a disposable DSN.
- Store conformance and explicit v0.93.0 upgrade/replay suite: passed with race
  instrumentation after adding the new table to partial-store fixture metadata.
- `go run ./cmd/sqlguard ./internal`: passed.
- Changed-package golangci-lint 2.9.0: passed with Go 1.26.0. An initial host
  Go 1.27 attempt was incompatible with the analyzer's export-data reader.
- Managed real-backend Chromium and Pixel-5/mobile-chrome scenarios: one test
  each passed, proving preview, Save/reload, invalid-variable error retention,
  restore/delete/reload and no horizontal phone scrolling. Synthetic editor
  screenshots are attached to the test reports. All seeded providers are disabled.
- Public-doc validator: 63 tests passed and 54 published pages validated.
- Full `make test` was attempted with two build workers but exceeded the
  600-second command budget during the broad backend suite. It is incomplete,
  not a passing full-suite claim; frontend/full-script phases were not reached.
- Full multi-project `make test-e2e` was not run. The two relevant managed
  desktop/phone scenarios above were executed sequentially with one shard/worker.

Final control/fuzz/stale-preview cases passed: the two mounted web suites now
contain 20 passing tests; the final Go 1.26 notification/config/weburl race run
passes. A three-second, two-worker fuzz run completed 12,759 executions without
exceeding rendering bounds. Production remains pending until the independent
rollout receipt confirms health and smoke results. That receipt now confirms
completion at 2026-11-03 10:08:40 UTC; see [operational evidence](deployment.md#results).
PR #5 is open; full-suite/time-budget and live PostgreSQL limitations above
remain explicit, so this delivery record does not claim all repository gates passed.

Design checks on 2026-11-03:

- `python3 scripts/list-docs.py validate`: passed (425 decisions, 1889 specifications).
- `python3 scripts/lint-spec-files.test.py`: passed (36 tests).
- `python3 scripts/lint-spec-files.py --all`: passed.
- Catalog query discovers both new specs. `git diff --check` and explicit
  `git diff --no-index --check /dev/null <file>` for all six new documents pass.
  Work-order REQ/AC references and design/plan links were reviewed against the
  paired requirement/design; all three work orders reference the plan's design.
- Those initial design checks do not substitute for the implementation evidence
  above. PR and production outcome are tracked separately in the rollout record.
- Live PR-documentation preflight requires a PR/event; it is not applicable to
  this uncommitted design handoff. No PR or artificial status was created.

Upstream proposal: [#5011](https://github.com/kdlbs/kandev/issues/5011), created,
updated with opaque-revision ABA protection, and read back. No comments or
maintainer discussion were present on read-back.
`git ls-remote origin refs/heads/main` still matched the baseline after posting.

## Contribution sequence

1. Open a feature issue following the repository form, with problem, proposal,
   affected boundaries, alternatives, compatibility/migration risks and this
   event matrix. Wait for maintainer discussion, not merely an issue number.
2. Reconcile this draft with that discussion and refreshed main. Start work only
   after the repository's subsequent implementation-request checkpoint.
3. Install workspace dependencies from `apps/` with `pnpm install --frozen-lockfile`.
   Run work-order checks and required `make fmt`, then `make typecheck test lint`,
   `make test-e2e`; record any pre-existing failures against the same baseline.
4. Inspect final status/diff/recent commits, review ownership and secrets paths,
   and commit only intended feature/docs/tests. Push the validated isolated
   branch to the authenticated user's fork, then create a PR to `kdlbs/kandev:main`.
5. PR body links the discussed issue and lists tests/manual evidence, screenshots,
   URL/missing-value policy, migration/restore and remaining limitations. Report
   the actual PR URL; no production rollout is part of this work.

## Risks and upstream decisions

- Maintainers must discuss Apprise-only scope, separate revisioned template table,
  origin config naming/policy, strict missing-value fallback and synthetic preview.
- A system-update template can reference optional subtype values; explicit help
  and preview must prevent a misleading impression that these always exist.
- Session-pair verification adds a bounded lookup; deletion races produce fallback,
  not replacement session IDs or relaxed ownership.
- Shared Save spans independent API transactions. Partial commits and uncertain
  outcomes need reconciliation; this proposal does not promise batch atomicity.
- Apprise CLI output currently enters errors; secure diagnostics must be covered
  in this feature before exposing template failures. OS process-list visibility
  of existing receiver CLI arguments is a pre-existing transport limitation,
  not solved by template redaction. Do not copy those arguments into new logs.
- PostgreSQL proof requires a disposable test DSN. If unavailable, explicitly
  report skipped live coverage rather than claim dialect parity was executed.
