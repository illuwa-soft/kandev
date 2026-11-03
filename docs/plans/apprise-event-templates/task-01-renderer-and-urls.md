---
id: "01-renderer-and-urls"
title: "Bounded catalog, renderer and URL contract"
status: done
wave: 1
depends_on: []
plan: "plan.md"
requirements:
  - REQ-PLATFORM-APPRISE-TEMPLATES-002
acceptance_criteria:
  - AC-PLATFORM-APPRISE-TEMPLATES-002.1
  - AC-PLATFORM-APPRISE-TEMPLATES-002.2
  - AC-PLATFORM-APPRISE-TEMPLATES-002.3
  - AC-PLATFORM-APPRISE-TEMPLATES-002.4
  - AC-PLATFORM-APPRISE-TEMPLATES-002.5
system_design:
  - ../../specs/platform/system-design/apprise-event-templates.md
---

# Task 01: Bounded catalog, renderer and URL contract

## Summary

Deliver a pure versioned placeholder parser/renderer, six-event catalog and
validated destination builders. Register the proposed optional operator URL
setting without changing existing delivery behavior.

## In scope

- Exact allowlists and conditional system-update shapes with synthetic examples.
- Limits, literal-only substitution, atomic pair fallback and structured diagnostics.
- Typed startup configuration/catalog/source and route/encoding parity fixtures.

## Out of scope

Persistent templates, sending, settings controls and live configuration changes.

## Acceptance

- Grammar/limits reject invalid data before rendering; output append is bounded.
- Every event/shape has truthful variables; missing values never synthesize a session.
- Origin and destinations pass malformed/scheme/encoding/unconfigured test matrices.

## Verification

Run failing tests before implementation and the same tests afterward. Proposed
new package paths below become runnable during this work order. From repository
root, after workspace dependency installation:

```bash
make fmt
```

From `apps/backend`:

```bash
go test -trimpath ./internal/notifications/templates/... ./internal/common/config/...
go test -trimpath -race ./internal/notifications/templates/...
```

## Files likely touched

- New `apps/backend/internal/notifications/templates/{catalog,renderer,urls}.go` and tests.
- `apps/backend/internal/common/config/{config,catalog,source}.go` and config tests.
- Existing Go/TypeScript link builders as parity references, without broad route refactors.

## Dependencies

Maintainer issue discussion, reconciled design, subsequent explicit implementation request.

## Risks

Escaping and URL encoding must be specified once; do not accidentally support
general template expressions or SPA subpath deployments.

## Parallelism

`sequential`

## Inputs

- Paired design: event catalog, grammar and public web URL sections.
- Existing `webpush.TaskURL`, `linkToTask` and typed config validation.

## Results

Implemented bounded literal parsing, allowlists, URLs and typed operator config.
Go 1.26.0 notification/config/weburl race suites and URL/grammar limit tests pass.
The compatibility base offers four event keys; newer upstream-only keys remain
outside this backport. See the plan for exact commands and toolchain evidence.
