---
id: "03-stable-backport"
title: "Verify a minimal stable-host backport"
status: pending
wave: 3
depends_on: ["02-consumer-contract"]
plan: "plan.md"
requirements:
  - REQ-PLUGINS-AUTOMATION-READS-001
acceptance_criteria:
  - AC-PLUGINS-AUTOMATION-READS-001.6
  - AC-PLUGINS-AUTOMATION-READS-001.8
  - AC-PLUGINS-AUTOMATION-READS-001.9
system_design:
  - ../../specs/plugins/system-design/automation-history-reads.md
---

# Task 03: Verify a Minimal Stable-Host Backport

## Summary

Prepare a separate compatibility branch at the exact installed stable source
and carry only the completed generic read extension. Prove the resulting
artifact against a disposable database before any operational replacement.

## In scope

- Baseline `v0.98.0`, commit `2d93502c256780e1f3d9df4348dafbd97b584188`.
- Existing backend SDK/proto layout, read adapters and corresponding tests.
- Existing frontend assets and runtime version/build identity preservation.

## Out of scope

- A wholesale main-branch upgrade, schema changes for a connector, or native
  agent/plugin configuration changes.

## Acceptance

1. The stable artifact contains only generic read functionality and its tests/docs.
2. Existing plugin wire behavior and read permissions pass on a disposable instance.
3. Baseline/candidate hashes, retained frontend assets, backup and rollback steps
   are recorded before deployment is considered ready.

## Verification

On the stable compatibility worktree, from its repository root:

```bash
go -C apps/backend test -race ./pkg/pluginsdk ./internal/plugins ./internal/automation
git diff --check v0.98.0
git diff --stat v0.98.0
```

Use the stable host's documented build target and execute readiness plus a
packaged independent-consumer smoke on a disposable instance. Record the exact
build/smoke command after inspecting that worktree's Makefile; do not reuse a
main-only build command or report an unrun production probe as evidence.

## Files likely touched

- Stable `apps/backend/pkg/pluginsdk` and `apps/backend/proto` equivalents.
- Stable Host read adapters and source integration tests.

## Dependencies

Tasks 01 and 02. This work order is a separate compatibility deliverable.

## Risks

The later main branch contains thousands of unrelated changes and a moved SDK.
Copying its runtime wholesale would exceed this read-only extension's scope.

## Parallelism

sequential

## Results

Pending.
