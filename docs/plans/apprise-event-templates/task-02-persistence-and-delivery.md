---
id: "02-persistence-and-delivery"
title: "Owned persistent APIs and Apprise delivery"
status: done
wave: 2
depends_on:
  - 01-renderer-and-urls
plan: "plan.md"
requirements:
  - REQ-PLATFORM-APPRISE-TEMPLATES-001
  - REQ-PLATFORM-APPRISE-TEMPLATES-002
  - REQ-PLATFORM-APPRISE-TEMPLATES-003
  - REQ-PLATFORM-APPRISE-TEMPLATES-004
acceptance_criteria:
  - AC-PLATFORM-APPRISE-TEMPLATES-001.1
  - AC-PLATFORM-APPRISE-TEMPLATES-001.2
  - AC-PLATFORM-APPRISE-TEMPLATES-001.3
  - AC-PLATFORM-APPRISE-TEMPLATES-001.4
  - AC-PLATFORM-APPRISE-TEMPLATES-002.2
  - AC-PLATFORM-APPRISE-TEMPLATES-002.4
  - AC-PLATFORM-APPRISE-TEMPLATES-002.5
  - AC-PLATFORM-APPRISE-TEMPLATES-003.2
  - AC-PLATFORM-APPRISE-TEMPLATES-004.1
  - AC-PLATFORM-APPRISE-TEMPLATES-004.2
system_design:
  - ../../specs/platform/system-design/apprise-event-templates.md
---

# Task 02: Owned persistent APIs and Apprise delivery

## Summary

Deliver revisioned title/body CRUD, synthetic preview and Apprise-only rendering
through every current notification entry point. Preserve subscription/recipient
and occurrence-claim contracts while proving migration and restart behavior.

## In scope

- Model/store/table, owner/type-checked CRUD/CAS, controller/DTO/HTTP endpoints.
- Migration replay, required-store descriptor/conformance and upgrade fixtures.
- Exact session-pair context, inbox context and all runtime/release shapes.
- Apprise CLI safe errors, no receiver/config/log leakage, side-effect-free preview.
- Full backup/restore and real PostgreSQL env-gated behavior tests.

## Out of scope

Web editor, new providers/events, real-data preview, transport redesign and deployments.

## Acceptance

- Foreign/missing targets are identical, CAS conflicts do not overwrite, provider
  removal/type change cleans overrides and restart/restore retain committed JSON.
- Every supported event uses custom/default pairs only for subscribed enabled
  Apprise providers; other providers, dedupe and fixed-copy test sends stay unchanged.
- Preview performs zero sends/history writes and diagnostics contain no receiver,
  token, source or rendered-message content, including CLI failure/timeout cases.

## Verification

Repository root:

```bash
make fmt
git diff --check
```

From `apps/backend`:

```bash
go test -trimpath -race ./internal/notifications/...
go test -trimpath ./internal/common/config/... ./internal/backendapp/...
go run ./cmd/sqlguard ./internal
go test -trimpath -race ./internal/persistence/storeconformance -count=1
```

Run the new integration tests once with `KANDEV_TEST_POSTGRES_DSN` pointing only
to a disposable DB supplied by approved test infrastructure. Do not print the
DSN. If unavailable, record live PostgreSQL coverage as skipped. Extend the
existing storeconformance explicit-tag upgrade matrix rather than add an
unrelated migration runner.

## Files likely touched

- `apps/backend/internal/notifications/{models,store,dto,controller,handlers,service}` and new template-focused tests.
- `apps/backend/internal/notifications/providers/apprise.go` and new `apprise_test.go`.
- `apps/backend/internal/backendapp/` notification wiring and tests.
- `apps/backend/internal/persistence/{requiredstores,storeconformance}` and upgrade fixtures.
- Existing backup/restore integration tests for a template-retention assertion.

## Dependencies

Task 01 plus maintainer discussion of the model/API and preview policy.

## Risks

Inbox bypasses the semantic dispatch helper today; runtime summaries retain
per-member claims. Integrate both without changing deduplication. Recheck
ownership/type within each SQL transaction, not just at an earlier API read.

## Parallelism

`sequential`

## Inputs

- Design: persistent model/API, delivery, migration, preview and security sections.
- Existing ownership, semantic-copy, atomic settings and runtime-summary tests.

## Results

Implemented owned revisioned CRUD, synthetic preview, schema/upgrade metadata,
Apprise-only dispatch and safe CLI diagnostics. Go 1.26.0 notifications/config
race tests, SQL guard and store conformance/explicit-tag replay pass. SQLite
file reopen and snapshot restore preserve JSON and revision. Live PostgreSQL
coverage is env-gated and was skipped without a disposable test DSN. The full
repository suite attempt timed out; no full-suite pass is claimed.
