---
id: "03-settings-and-e2e"
title: "Settings editor, documentation and E2E"
status: in_progress
wave: 3
depends_on:
  - 02-persistence-and-delivery
plan: "plan.md"
requirements:
  - REQ-PLATFORM-APPRISE-TEMPLATES-001
  - REQ-PLATFORM-APPRISE-TEMPLATES-002
  - REQ-PLATFORM-APPRISE-TEMPLATES-003
  - REQ-PLATFORM-APPRISE-TEMPLATES-004
acceptance_criteria:
  - AC-PLATFORM-APPRISE-TEMPLATES-001.1
  - AC-PLATFORM-APPRISE-TEMPLATES-001.2
  - AC-PLATFORM-APPRISE-TEMPLATES-002.1
  - AC-PLATFORM-APPRISE-TEMPLATES-002.3
  - AC-PLATFORM-APPRISE-TEMPLATES-003.1
  - AC-PLATFORM-APPRISE-TEMPLATES-003.2
  - AC-PLATFORM-APPRISE-TEMPLATES-003.3
  - AC-PLATFORM-APPRISE-TEMPLATES-003.4
  - AC-PLATFORM-APPRISE-TEMPLATES-004.3
system_design:
  - ../../specs/platform/system-design/apprise-event-templates.md
---

# Task 03: Settings editor, documentation and E2E

## Summary

Expose the delivered template API through the existing Notifications settings
owner and shared Save coordinator. Add public guidance and desktop/phone browser
proof using only synthetic/disposable data.

## In scope

- API client/types, store/hook, provider/event editor, catalog/preview and errors.
- Staged restore, shared Save/Reset, held-request reconciliation and CAS conflicts.
- All required translations, settings discovery, keyboard/touch accessibility.
- Public setup, variables/defaults, URL startup configuration, API errors and
  database/config backup/restore instructions.
- E2E, manual checks/screenshots and accurate eventual PR validation report.

## Out of scope

Production changes, unrelated settings refactors and all-provider templates.

## Acceptance

- Users complete create/edit/delete/restore/preview through shared Save/Reset,
  with separate subscription status and retained drafts/errors/later edits.
- Desktop and phone meet UI-01/UI-02 semantics and show synthetic-only previews;
  tests assert zero sends and protect against stale preview publication.
- Public examples match real API/config contracts and documented verification
  results include all skipped/unavailable checks.

## ASCII UI preview

UI-01 excerpt, shared desktop/phone composition; see the
[full views and UI-02 states](plan.md#ascii-ui-preview):

```text
Apprise message templates
  Provider [v]  Event [v]  Subscription: Off  Content: Custom
  Title [{{task.title}}]
  Body  [{{default.body}}]
        [{{session.url}}]
  Variables [insert/help]  [Restore defaults]
  Sample [v] [Preview] -> Synthetic result / fallback warning
  [shared Reset] [shared Save changes]
```

Maps to `AC-PLATFORM-APPRISE-TEMPLATES-003.1` through `.4`. Stack selects on
phone; wrap variables; keep coarse-pointer targets at least 44px and avoid
horizontal scrolling. Use translated error/status regions, not button text alone.

## Verification

From `apps/`, before any frontend commands:

```bash
pnpm install --frozen-lockfile
```

From `apps/web` (new test files are delivered by this work order):

```bash
pnpm exec vitest run components/settings/apprise-template-editor.test.tsx components/settings/notification-save-drafts.test.tsx hooks/domains/settings/use-notification-providers.test.tsx
pnpm run typecheck
pnpm run i18n:check
pnpm e2e:run --host --shards 1 --project chromium tests/settings/apprise-event-templates.spec.ts
pnpm e2e:run --host --shards 1 --project mobile-chrome tests/settings/mobile-apprise-event-templates.spec.ts
```

Repository-root contribution gates after final edits:

```bash
make fmt
make typecheck test lint
make test-e2e
python3 scripts/list-docs.py validate
python3 scripts/lint-spec-files.test.py
python3 scripts/lint-spec-files.py --all
git diff --check
```

Run the public-doc validation and PR-doc coverage preflight documented by the
repository at implementation time. Do not invent a PR event or publish a status
to satisfy local preflight. Inspect helpers' CLI contracts first. Capture manual
desktop/phone screenshots from the isolated test backend, not the live instance.

## Files likely touched

- `apps/web/components/settings/notifications-settings*.tsx` and new small editor components/tests.
- `apps/web/hooks/domains/settings/`, `lib/api/domains/`, settings store and HTTP types.
- `apps/web/src/locales/`, settings discovery catalog.
- New `apps/web/e2e/tests/settings/{apprise-event-templates,mobile-apprise-event-templates}.spec.ts` and shared helpers.
- `docs/public/` relevant settings/config/API/data-management guides (inspect public-doc contribution guide first).
- This plan and paired specifications for implemented status/results only after proof.

## Dependencies

Task 02 API and tested dispatch behavior. No frontend duplicate renderer.

## Risks

Current settings save preserves drafts edited during pending requests; template
drafts must join that owner without losing this guarantee. Catalog/example data
must not carry provider config. E2E must use guarded shard settings and causal waits.

## Parallelism

`sequential`

## Inputs

- UI-01/UI-02 and design settings/preview/API sections.
- Existing notification-save-drafts tests/helpers and shared save coordinator.

## Results

Editor/API/hook integration and public documentation implemented. Shared Save,
restore/delete and later-edit tests pass; Node 24 typecheck/i18n and scoped
ESLint pass. One managed real-backend Chromium test and one native Pixel-5
mobile-chrome test pass. Public-doc validator tests (63) and 54 published pages
pass. Final stale-preview/control checks pass (20 mounted tests total), and
[PR #5](https://github.com/illuwa-soft/kandev/pull/5) is open. The exact production
artifact passed isolated desktop/phone tests and the rollout receipt confirms
operational CRUD/preview/fallback with no test sends. Full all-project E2E and
the complete broad repository gate remain unverified as recorded in the plan;
this work order stays in progress for that validation remainder.
