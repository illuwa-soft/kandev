---
status: current
system: plugins
requirements:
  - REQ-PLUGINS-AUTOMATION-READS-001
---

# Plugin Automation History Read Design

## Purpose and boundaries

This additive Host read extension is implemented in source; its release and
stable-backport verification remain separate delivery results.
The Host projects existing automation service data into stable public DTOs.
The extension owns authorization, projection, pagination, and capability
discovery. It does not own automation execution or any consumer's business logic.

At source revision `dc2d77359f101fe398255112346b60c617477162`, the standalone
`pluginsdk` contains managed-conversation schedule methods, but no general
automation-run history reader. The application exposes `automation.runs.list`
and retry history through its own transport. These application routes are
evidence of existing data, not a plugin API to call from a connector.

## Requirement mapping

| Criteria | Design boundary                                     |
| -------- | --------------------------------------------------- |
| .1-.3    | Public observations and existing service projection |
| .4-.5    | Bounded snapshots and exact refresh                 |
| .6-.7    | Workspace authorization and secret-free DTOs        |
| .8-.9    | Read-only dependency boundary and compatibility     |

All entries refer to `AC-PLUGINS-AUTOMATION-READS-001`.

## Components and responsibilities

- `pluginsdk/proto/kandev/plugin/v1/plugin.proto`: additive read RPCs and DTOs.
- `pluginsdk`: optional Go reader extension and gRPC adapters. Helper:
  `AutomationHistory(host)`, returning a reader or an unsupported result.
- `apps/backend/internal/plugins`: capability checks, exact workspace scope,
  bounded read snapshots, DTO mapping, and read-method discovery.
- `apps/backend/internal/automation`: the existing service owns automation and
  run reads. Reuse its run paging and retry identity; do not query its store from
  a plugin or introduce a connector-specific branch in this package.

Reader operations are `ListAutomations`, `ListRuns`, and `GetRun`. Their additive
wire methods are `ListAutomationDefinitionsExact`, `ListAutomationRunsExact`, and
`GetAutomationRunExact`. Source implementation does not establish a published
SDK release or a compatible installed Host.

## Public observations

An automation observation includes identity, workspace, display name, enabled
state, safe schedule/timezone and trigger-kind projections, source creation and
update times, resource version, and observation time.

A run observation includes identity and workspace; automation/trigger identity;
source creation time; task/session/turn identities when present; original status,
summary and error projection; retry-group, parent run and attempt number; and a
projection version and observation time. Exclude raw trigger data and credential
or claim fields. Preserve opaque IDs as strings and timestamps as RFC3339.
Keep taskless and skipped executions representable without invented task IDs.

Do not add domain timestamps merely to make the connector's cursor easier.
Where the source does not retain a transition time or revision, return explicit
absence and a version of the observed projection. Observation time is not an
execution timestamp.

## Bounded snapshots and exact refresh

Reuse the existing Host exact-query snapshot and receipt conventions where
possible. A page is bound to installation, approval revision, workspace, query,
and snapshot identity. Publish a bounded next cursor, exhaustion flag, and
observation time. Invalidated snapshots require a fresh query rather than a
fabricated continuation. Use deterministic source ordering with identity as a
tie-breaker.

An initial bounded scan discovers identities. A consumer can re-read known runs
with `GetRun`, including runs admitted before the consumer started. This keeps
outcome reconciliation independent of creation-time cursors. A list snapshot
does not promise to contain later source updates. Deletion and retention remain
owned by the source; a no-longer-visible record is not inferred to have succeeded.

## Authorization and failures

Require `api_read: automations`, matching workspace approval such as
`host.v2.read:automations`, and current method support. Reuse existing Host
capability context and authorization conventions; do not let a plugin claim
grant identity. Validate an automation or run belongs to the authorized
workspace before returning it. Recheck revocation on every continuation.

Use explicit permission, unsupported, cursor-expired, not-visible, and upstream
read errors. An error is not an empty inventory. Reads make no durable writes.
The DTO mapper is an allowlist; it must not forward a domain struct or arbitrary
configuration/body JSON through a protobuf Struct.

## Compatibility and publication

Preserve the existing wire service and add optional methods. An older Host must
return unsupported and a new SDK must remain usable for its existing methods.
The standalone SDK move has no published host mapping in the inspected source;
verify its public release record before changing a consumer's dependency.

The operator baseline is release `v0.98.0` at
`2d93502c256780e1f3d9df4348dafbd97b584188`. It predates the standalone SDK move.
If that release needs a local backport, carry only this read contract through
the existing backend-owned SDK/proto paths. Do not deploy the whole later main
branch as an incidental prerequisite. A manifest version floor alone is not
proof that a backported method is available; require method capability evidence.

## Verification and observability

Cover permission denial before dependencies, foreign identity, missing grants,
revocation between pages, exhausted/expired/mismatched cursors, source status
changes, taskless/skipped runs, retry identity, secret omission, old-Host
fallback, and the gRPC round trip. Assert source state is unchanged after reads.
Use real service/store fixtures for data integration and both SQLite and the
existing environment-gated PostgreSQL lane for changed dialect-sensitive reads.

Log bounded method, result, latency, and request/receipt identity without content
or secrets. Do not add a scheduler, a connector worker, or an external API client
to the Host extension.
