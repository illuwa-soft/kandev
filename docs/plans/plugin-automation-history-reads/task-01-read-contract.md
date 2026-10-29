---
id: "01-read-contract"
title: "Add authorized automation history reads"
status: completed
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

- Added the optional public reader, additive protobuf messages/RPCs, generated
  adapters, workspace authorization, capability discovery, and live service wiring.
- Reused immutable exact-query snapshots and the automation service's paged reads.
  Exact run refresh now shares list-time status and exact-turn summary projection.
- RED observed for unavailable Host methods, missing discovery entries, and a raw
  `task_created` result where the source projection required `archived`.
- Passed `go -C pluginsdk test -trimpath -race ./...` and the full affected backend
  package tests with `KANDEV_PLUGIN_FIXTURE_TEST_LOOPBACK=1` and `-trimpath -race`.
  Without that existing fixture setting, the unmodified baseline and candidate
  both time out on the host's selected VPN address. No test was disabled.
- PostgreSQL parity coverage is present and environment-gated; no PostgreSQL DSN
  was supplied for this local run.
- Changed-package golangci-lint v2.9.0 passed with Go 1.26.0, `-trimpath`, four
  workers, and a 15-minute cold-cache deadline. The installed Go 1.27 export format
  is newer than that pinned linter's importer supports.
- `git diff --check` passed. No source write, automation dispatch, or vendor call
  was added to this Host extension.
