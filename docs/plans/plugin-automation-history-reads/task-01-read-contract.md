---
id: "01-read-contract"
title: "Add authorized automation history reads"
status: pending
wave: 1
depends_on: []
plan: "plan.md"
requirements:
  - REQ-PLUGINS-AUTOMATION-READS-001
acceptance_criteria:
  - AC-PLUGINS-AUTOMATION-READS-001.1
  - AC-PLUGINS-AUTOMATION-READS-001.2
  - AC-PLUGINS-AUTOMATION-READS-001.3
  - AC-PLUGINS-AUTOMATION-READS-001.4
  - AC-PLUGINS-AUTOMATION-READS-001.5
  - AC-PLUGINS-AUTOMATION-READS-001.6
  - AC-PLUGINS-AUTOMATION-READS-001.7
  - AC-PLUGINS-AUTOMATION-READS-001.8
system_design:
  - ../../specs/plugins/system-design/automation-history-reads.md
---

# Task 01: Add Authorized Automation History Reads

## Summary

Implement one vertical read-only Host extension from public SDK through existing
automation services, including capability discovery and safe DTO projection.

## In scope

- Proposed SDK reader, additive proto methods, generated adapters, Host wiring.
- Real service fixtures for paging, updated outcomes, and unchanged source data.
- Denial-before-dependency, revocation, secret omission, and malformed cursors.

## Out of scope

- Consumer policy, provider calls, database access from plugins, and UI changes.

## Acceptance

1. Authorized reads preserve source identities/outcomes across bounded pages.
2. Missing/revoked grants, foreign identities, and invalid cursors fail explicitly.
3. Reads leave automation/task/execution state unchanged and omit all secret fields.

## Verification

From the repository root, after authoring the focused tests and regenerating proto:

```bash
go -C pluginsdk test -race ./...
go -C apps/backend test -race ./internal/plugins ./internal/automation
git diff --check
```

Use `automation_history_test.go` in the SDK for the wire round trip and
`host_automation_history_test.go` in the Host for source integration, foreign
scope, secret omission and revocation. Names are planned new files.

## Files likely touched

- `pluginsdk/automation_history.go` and its focused tests (new).
- `pluginsdk/proto/kandev/plugin/v1/plugin.proto` and generated outputs.
- `apps/backend/internal/plugins/host_automation_history.go` (new).
- Existing exact capability registry, provider wiring and automation read service.

## Dependencies

None. An explicit implementation request after the design handoff is required.

## Risks

The source has separate creation and completion observations. Do not flatten
unknown/accepted state into success or infer missing times.

## Parallelism

sequential

## Results

Pending.
