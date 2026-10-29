---
status: active
system: plugins
created: 2026-10-29
owners:
  - kandev
---

# Plugin Automation History Reads

## Overview

Plugins need an authorized, typed way to reconcile automation definitions and
execution outcomes after missing a live event or restarting. Reporting, audit,
calendar, and memory connectors share this requirement. The plugin system owns
the public read contract; automation scheduling and execution remain with their
existing owners.

## Requirements

### REQ-PLUGINS-AUTOMATION-READS-001: Workspace-scoped history reads

**Intent:** A plugin can discover automations and read their execution history
without an internal database, application HTTP client, or scheduling policy.

#### Acceptance criteria

- **AC-PLUGINS-AUTOMATION-READS-001.1:** An authorized plugin shall be able to list
  automation definitions and execution records in an explicitly selected
  workspace and retrieve a known execution by its identity.
- **AC-PLUGINS-AUTOMATION-READS-001.2:** Each execution shall preserve its source
  automation, trigger, task, session, turn, retry-group, parent-attempt, and
  attempt-number identities when present. Absent identities shall remain absent.
- **AC-PLUGINS-AUTOMATION-READS-001.3:** Reads shall expose source timestamps,
  outcome, summary, and error projections without presenting a scheduled,
  accepted, skipped, or unknown execution as successfully completed.
- **AC-PLUGINS-AUTOMATION-READS-001.4:** A consumer shall be able to exhaust bounded
  pages without silently omitting records or treating a partial page as a full
  inventory. Cursors shall be opaque and bound to the original workspace and
  query. Expired or mismatched cursors shall fail explicitly.
- **AC-PLUGINS-AUTOMATION-READS-001.5:** An execution whose outcome changes after a
  list observation shall be readable again by identity. Reconciliation shall
  not depend on delivery of its original event.
- **AC-PLUGINS-AUTOMATION-READS-001.6:** The Host shall require the declared read
  capability and the applicable workspace grant before reading automation data.
  Foreign-workspace and nonexistent identities shall have equivalent not-visible
  outcomes. Revocation shall prevent subsequent pages and exact reads.
- **AC-PLUGINS-AUTOMATION-READS-001.7:** Returned records shall omit credentials,
  webhook secrets, request headers, raw trigger bodies, and internal claim tokens.
  Public DTOs shall not be serialized internal domain models.
- **AC-PLUGINS-AUTOMATION-READS-001.8:** Reads shall not trigger work, retry an
  execution, change an automation, create a task, run an agent, or mutate domain
  state. Vendor-specific analysis and export policy shall remain with consumers.
- **AC-PLUGINS-AUTOMATION-READS-001.9:** Hosts without the extension shall return an
  explicit unsupported capability. Existing plugins and managed-conversation
  schedule methods shall retain their existing contracts.

## Out of scope

- Provider-specific memory, calendar, summarization, retention, and bank policy.
- A durable event journal, workflow automation changes, or an Office dependency.
- A global prompt-transform hook or new mutation capability.
- A declaration that an unimplemented read surface is available in a release.

## Related contracts

- [Plugin system](../README.md).
- [Design](../system-design/automation-history-reads.md).
- [Automation runs](../../office/requirements/automation-runs.md).
