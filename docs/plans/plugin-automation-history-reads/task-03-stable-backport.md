---
id: "03-stable-backport"
title: "Verify a minimal stable-host backport"
status: completed
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

- Backported only the generic read extension to `v0.98.0` at
  `2d93502c256780e1f3d9df4348dafbd97b584188`, using the backend-owned SDK/proto layout.
  No schema, agent-control, frontend-source, or provider-specific changes were added.
- Passed the full SDK, plugin, and automation package race tests with
  `KANDEV_PLUGIN_FIXTURE_TEST_LOOPBACK=1`, `-trimpath`, and four build workers.
- Rebuilt the unchanged stable frontend with Node 24.15.0, pnpm 9.15.9 and the
  frozen stable lockfile, then ran `make sync-embedded-web`. A backend-only build's
  fallback shell is not a deployment artifact.
- Built with `make -C apps/backend build-kandev
  VERSION=0.98.0+automation-history.1 GOFLAGS=-p=4`; retained the original stable
  helper contract rather than importing the later main runtime.
- On a disposable home/database, readiness reported the candidate version, a
  separately compiled current-SDK consumer package installed, an unapproved read
  returned 403, and an approved read returned a bounded empty inventory/receipt.
  The bundled SPA entry module was fetched successfully.
- Copied the stopped disposable state independently of the source, checked SQLite
  integrity, and restored it under the original Homebrew v0.98.0 native executable.
  The workspace survived, readiness returned v0.98.0, and the new reader returned
  explicit unsupported (501 from the probe). All owned processes were stopped.
- Original native executable SHA-256:
  `31b99ddf3a0f63370edfc88482abefba3201f4ce0d196b50bb31449eb780df84`.
  Final candidate/packaged artifact hashes are delivery receipts, recorded after
  the final source commit. No production runtime was replaced by this work order.
- Public documentation/catalog/spec validation and `git diff --check v0.98.0`
  passed. `git diff --exit-code v0.98.0 -- apps/web apps/packages` confirmed the
  original frontend and shared-package source was retained.
