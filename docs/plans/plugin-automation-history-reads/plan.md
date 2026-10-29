---
created: 2026-10-29
status: completed
requirements:
  - REQ-PLUGINS-AUTOMATION-READS-001
system_design:
  - ../../specs/plugins/system-design/automation-history-reads.md
legacy_specs: []
---

# Implementation Plan: Plugin Automation History Reads

## Overview

Add a generic read contract for automation definitions and execution history.
The plugin system owns the contract; consumers own export and analysis policy.
The source contract and minimal stable backport are implemented and verified.
Runtime installation remains a separate operator delivery step.

## Scope

### In scope

- Additive SDK/wire contract, Host service projection, authorization and paging.
- Focused gRPC/service integration tests and public author documentation.
- An independently verified minimal backport for an existing stable installation.

### Out of scope

- Memory/calendar provider logic, execution mutations, prompts, and schedulers.
- Upgrading an installation to all changes on main to obtain one read method.

## Technical approach

Follow the [design](../../specs/plugins/system-design/automation-history-reads.md).
Read the scoped agent instructions and backend test guide before implementation.
Reuse the existing exact-read and service boundaries; keep source and generated
proto artifacts synchronized through the repository generation target.

| Host                    | SDK location            | Required result                                    |
| ----------------------- | ----------------------- | -------------------------------------------------- |
| Current main contract   | root `pluginsdk` module | Optional reads with current workspace approvals    |
| Original v0.98.0        | backend `pkg/pluginsdk` | Existing methods unchanged; new reader unsupported |
| Minimal stable backport | backend-owned SDK/proto | Same read DTO semantics, independently tested      |

## Tests

Task 01 owns wire/projection, authorization, paging and read-only integration.
Task 02 owns public contract validation and old-Host consumer evidence.
Task 03 owns stable backport compatibility and artifact identity. All nine
acceptance criteria are assigned in the work orders.

## Work orders

- [x] [Task 01: Read contract](task-01-read-contract.md)
- [x] [Task 02: Consumer contract](task-02-consumer-contract.md)
- [x] [Task 03: Stable backport](task-03-stable-backport.md)

## Verification results

Tasks 01 and 02 shipped to the personal fork in PR #2. Task 03 records stable
package tests, a live packaged-consumer smoke, bundled frontend verification,
and restoration under the original v0.98.0 executable. These results do not claim
a production installation or public SDK release.

## Risks

- A declaration without an authorized backing data reader must remain unsupported.
- Creation-time pagination alone does not discover updates to an existing run.
- A stable backport uses a different source layout from the extracted SDK.
- Event delivery is best-effort; it is not the reconciliation inventory.
