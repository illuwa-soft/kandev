---
id: "02-consumer-contract"
title: "Document and prove the optional consumer contract"
status: completed
wave: 2
depends_on: ["01-read-contract"]
plan: "plan.md"
requirements:
  - REQ-PLUGINS-AUTOMATION-READS-001
acceptance_criteria:
  - AC-PLUGINS-AUTOMATION-READS-001.4
  - AC-PLUGINS-AUTOMATION-READS-001.6
  - AC-PLUGINS-AUTOMATION-READS-001.9
system_design:
  - ../../specs/plugins/system-design/automation-history-reads.md
---

# Task 02: Document and Prove the Optional Consumer Contract

## Summary

Add the reader to the authoritative author matrices and prove old-Host fallback
with an independent consumer. Keep all example behavior provider-neutral.

## In scope

- Public authoring/manifest and wire-contract documentation.
- Optional-method fallback and capability-discovery consumer tests.

## Out of scope

- Publishing an unverified SDK release or a product connector.

## Acceptance

1. Documentation matches the implemented signatures and workspace approvals.
2. Old-Host use reports unsupported without an internal transport fallback.
3. Paging and grant failures remain explicit to an independently compiled caller.

## Verification

```bash
go -C pluginsdk test -race ./...
node --test scripts/validate-public-docs.test.mjs
node scripts/validate-public-docs.mjs
python3 scripts/list-docs.py validate
python3 scripts/lint-spec-files.py --all
git diff --check
```

## Files likely touched

- `docs/public/plugins-authoring.md` and `docs/public/plugins-manifest.md`.
- `docs/plans/plugins/GRPC-CONTRACT.md`.
- SDK consumer/compatibility tests and maintained fixture.

## Dependencies

Task 01.

## Risks

A version string or schema declaration is not evidence of a functioning reader.

## Parallelism

sequential

## Results

- Updated the public authoring/manifest references and the wire contract with
  implemented method names, grant requirements, snapshot bounds, and old-Host errors.
- Added a separately compiled consumer executable under SDK testdata. Its packaged
  go-plugin handshake exercised the real Host reader and preserved paging mismatch,
  revoked-grant, and unsupported-Host outcomes without a private transport fallback.
- SDK race tests and `TestAutomationHistoryIndependentConsumer` passed.
- Both public-doc validators, documentation catalog validation, full specification
  lint, and `git diff --check` passed.
- No public SDK version or Host release mapping was published or inferred.
