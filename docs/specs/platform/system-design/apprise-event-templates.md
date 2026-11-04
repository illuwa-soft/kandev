---
status: draft
system: platform
requirements:
  - REQ-PLATFORM-APPRISE-TEMPLATES-001
  - REQ-PLATFORM-APPRISE-TEMPLATES-002
  - REQ-PLATFORM-APPRISE-TEMPLATES-003
  - REQ-PLATFORM-APPRISE-TEMPLATES-004
---

# Apprise Event Templates System Design

## Compatibility delivery amendment

The owner requested implementation on the exact Mac runtime after the original
upstream investigation. Delivery base is `production/mac` at
`2d8aee5817645b4e8518e56d5d268f0addac1c79` (existing fork PR #4), matching
the installed `0.98.0+upstream-automation-history.1` receipt and binary SHA-256
`17b6c0726bc4c563e0a6a5a140b0fb81fcb250d5f2008f18a602219051e9bb02`.
The two newer upstream session events below are research context only and are
excluded from this implementation. Its catalog exposes the four keys returned
by this runtime's `AvailableEvents`. Web Push and Webhook are not present on
this base and are not introduced. The initial evidence table below describes
the investigated upstream revision, not line numbers on the compatibility base.

Implementation uses `templates.go`, `template_values.go` and `template_catalog.go`
in the existing notification service; the new `notification_templates` table is
part of the existing required-store descriptor and explicit-tag upgrade manifest.
The public origin is parsed by neutral `internal/common/weburl` and registered
in the startup config catalog/source. UI catalog values are arrays of allowed
names and samples; conditional help is localized in the editor and fully defined
in the public guide. Validation errors use bounded generic codes rather than
echoing unknown names/source offsets. Four editors/catalog records replace the
original six-record API proposal. These are intentional scope adaptations.

The fork's disabled issue tracker and owner's explicit proceed/deploy request
are recorded here as the fork-local direction. Do not claim upstream maintainer
discussion or submit an upstream PR without it.

## Context and evidence

Baseline: upstream `main` at `c060076d6723a611ebc260371aa80b555442e589`.
All additions below are proposed, not existing APIs or shipped settings.
Platform owns notification data, delivery and settings, so this remains one
vertical requirement/design pair.

| Existing source | Observed contract and design consequence |
| --- | --- |
| `apps/backend/internal/notifications/service/service.go:23-31,103-111` | Six event keys. Do not introduce a HITL-specific subscription. |
| Same file, `deliverOccurrence:333-361` | Enabled/subscription checks precede delivery claim and dispatch. Keep that order. |
| Same file, `HandleInboxItem:456-497` | Separate generic dispatch, no task/session/item ID; must join the customization seam without inventing context. |
| Same file, `semanticMessageCopy:554-580` | Existing defaults and missing-title variants are the fallback source. |
| `service/runtime_updates.go:15-38,68-157` | System update also covers single runtime outcomes and grouped available updates; preserve per-member claims. |
| `providers/apprise.go:27-46` | CLI receives separate title/body arguments; raw CombinedOutput currently reaches an error. Sanitize this path before adding user content. |
| `store/sqlite.go:51-108,303-317,378-385` | Separate subscriptions and provider-owner predicates; initialization supports replay and PostgreSQL. |
| `handlers/handlers.go:21-32`, `controller/controller.go:23-33` | Existing provider CRUD/test routes and request-derived user identity. |
| `apps/web/components/settings/notifications-settings.tsx:160-255` | Shared Save/Reset owner and independent events table. Extend these, not an immediate-save parallel form. |
| `apps/web/lib/links.ts:7-20`, `webpush/payload.go:40-51` | Canonical `/t/{task}?sessionId={session}` route, encoding each component separately. |
| `apps/web/src/settings-routes.tsx:232`, `src/office-routes.tsx:58` | Current update and inbox destinations are `/settings/system/updates` and `/office/inbox`. |
| `internal/common/config/config.go:155-181` | Broker public URL has a separate purpose; reuse validation policy, not that setting. |
| `docs/specs/platform/system-design/notification-settings-atomicity.md` | Provider writes are atomic individually; UI batch Save is not globally atomic. |

Existing tests to extend: `notifications/handlers/ownership_test.go`,
`store/sqlite_test.go`, `service/service_test.go`,
`service/runtime_update_summary_test.go`, frontend
`components/settings/notification-save-drafts.test.tsx` and
`e2e/tests/settings/notification-save-drafts-helpers.ts`.

Related merged work: [#1918](https://github.com/kdlbs/kandev/pull/1918)
(semantic events), [#3551](https://github.com/kdlbs/kandev/pull/3551)
(Apprise rescan), [#4683](https://github.com/kdlbs/kandev/pull/4683)
(save races), [#4899](https://github.com/kdlbs/kandev/pull/4899)
(Webhook). Searches for Apprise, notification templates and template/Apprise
across issues and PRs found no matching customization proposal at investigation
time. Search results are not proof that no unindexed/private work exists.

## Requirement mapping

| Requirement | Sections |
| --- | --- |
| `REQ-PLATFORM-APPRISE-TEMPLATES-001` | Scope; storage/API; delivery |
| `REQ-PLATFORM-APPRISE-TEMPLATES-002` | Event catalog; grammar; URL contract |
| `REQ-PLATFORM-APPRISE-TEMPLATES-003` | Preview; settings |
| `REQ-PLATFORM-APPRISE-TEMPLATES-004` | Storage/migrations; security; verification |

## Scope and alternatives

Apply overrides only to Apprise. A small pure renderer/catalog under
`internal/notifications` can be independently tested, but only the Apprise
dispatch branch consumes its results. Do not add a general template plugin
system or change other providers' content contracts. Generic all-provider
templates would risk Local localized update rendering, native notifications,
Web Push size limits and the versioned Webhook payload. Storing overrides in
the opaque credential-bearing provider `config` would couple ordinary URL edits
to template replacement and expose too much data to preview. A separate table
and narrowly scoped APIs avoid both problems.

The existing sending `/test` operation remains fixed-copy and explicit. Preview
does not call it. Unsubscribe retains an override for later reuse; provider
deletion removes it. A provider type change away from Apprise deletes its
overrides in the same provider transaction to prevent dormant resurrection.

## Event catalog and variable matrix

All events support `event.type`, `default.title`, `default.body`. The latter two
are the already-built default pair for this occurrence, not recursively rendered
templates. `event.url` is conditional on destination availability. All other
variables below are an allowlist, not reflection over event payload/config.

| Event key | Additional variables | Source and destination |
| --- | --- | --- |
| `session.turn_finished` | `task.id`, `task.title`, `task.url`, `session.id`, `session.url`, `event.url` | Supplied task/session/turn occurrence; task title from that task; event URL equals the exact session URL. |
| `session.clarification_requested` | Same session variables | Supplied task/session/pending occurrence; no question text; event URL equals session URL. |
| `session.permission_requested` (newer upstream only) | Same session variables | Not offered on this compatibility base; no new event source is added. |
| `session.failed` (newer upstream only) | Same session variables | Not offered on this compatibility base; no new event source is added. |
| `office.inbox_item` | `inbox.title`, `inbox.type`, `event.url` | Existing workspace/item-type/title input; `/office/inbox` with source workspace encoded as `workspaceId`; no item ID is currently provided. |
| `system.update_available` | `update.kind`, `update.version`, `update.previous_version`, `update.agent_name`, `update.runtime_id`, `update.status`, `update.count`, `release.url`, `event.url` | Per-shape availability below; never task/session variables. |

| Update shape | Available subtype values | Destination |
| --- | --- | --- |
| Kandev release | `update.kind=kandev`, `update.version`, validated `release.url` when supplied | `event.url` is configured web origin + `/settings/system/updates`; release URL is a separate optional HTTPS external link. |
| Single runtime available/succeeded/failed/interrupted | `update.kind=agent_runtime`, version, previous_version, agent_name (display name), runtime_id, status, count=1 | `/settings/agents#runtime-update-{encoded agent ID}` from known runtime context; build locally instead of trusting arbitrary payload URL. |
| Grouped available runtimes | `update.kind=agent_runtime_summary`, `update.status=available`, `update.count` | `/settings/agents#runtime-updates`; no single version, agent name or runtime ID. |

Subtype-inapplicable update values remain catalogued as conditional and absent
at runtime. The editor/preview states those conditions. Do not create new
subscription keys or one template per subtype in this change.

### Default copy reference

These are existing behavior, not new default templates. Keep code as the single
default-copy owner; catalog responses render its synthetic examples.

| Event/shape | Title | Body |
| --- | --- | --- |
| Turn | `Agent turn finished` | `The agent finished a turn on "<title>".`; without title: `The agent finished a turn.` |
| Clarification | `Agent needs your answer` | `The agent asked a question on "<title>".`; without title: `The agent asked a question.` |
| Permission | `Agent needs permission` | `The agent is waiting for your approval on "<title>".`; without title: `The agent is waiting for your approval.` |
| Failed | `Agent session failed` | `The agent session failed on "<title>".`; without title: `The agent session ended with an error.` |
| Inbox | `Inbox: <type>` or `New inbox item` | Input item title |
| Release | `Kandev update available` | `Kandev <version> is available. Open Settings > System > Updates to review it.` |
| Runtime summary | `<count> agent runtime updates available` | `Review runtime versions in Settings > Agents.` |
| Single runtime | `<name> runtime update available/updated/update failed/update interrupted` | Preserve the full status-specific output of `runtimeUpdateMessage` at `runtime_updates.go:146-157`; test exact output for each status rather than copy a second formatter. |

## Grammar, limits and missing values

Version 1 grammar: literal text plus `{{variable.name}}`, with optional ASCII
space around the name. Names are exact, case-sensitive catalog keys. Escape
literal braces with `\{` and `\}` and backslash with `\\`; reject all other
backslash escapes, unmatched braces and nested placeholders. No functions,
operators, filters, loops, indexing, evaluation, includes or raw payload access.
Do not use Go `text/template` with exposed methods or an expression engine.

Parse once into bounded literal/key tokens. Validate source sizes (title 1..1024
bytes, body 1..16384), UTF-8, NUL, title CR/LF, and at most 64 placeholders per
field. Bound output before each append (2048-byte title, 65536-byte body), including
escaped-literal expansion and substituted values. Revalidate title controls after
substitution. The single-pass renderer never parses a value that contains braces.

Unknown keys are write/preview errors. A known-but-unavailable value, invalid
saved syntax/version, read error or output-limit violation at delivery uses the
entire existing default pair; never mix a custom title with a default body.
Empty task title is unavailable rather than a made-up label. Empty session ID
means both session fields and session event destination are unavailable, even
if a task URL could be built. Independent `task.url` can still be used.
This strict policy avoids silently producing broken links; synthetic previews
make it visible before save. Constant-only templates still render when URLs
are unconfigured. Saving a syntactically valid conditional variable is permitted.

## Public web URL contract

Propose `notifications.publicBaseUrl` with environment alias
`KANDEV_NOTIFICATIONS_PUBLIC_BASE_URL`, default empty. Register in the existing
typed config catalog/source, document precedence (environment > selected YAML >
default), and inject the validated value at service construction. It is an
operator startup setting requiring restart, not a user-supplied callback. Do not
reuse `githubCredentialBroker.publicBaseUrl` or infer from Host/Forwarded headers.

Accept an absolute origin with optional trailing `/`, valid nonempty host/port,
HTTPS, or HTTP only for existing loopback-host policy. Reject opaque/network-only
URLs, userinfo, query (including bare `?`), fragment (including bare `#`), control
characters, dot/path segments, non-root paths and malformed escapes. Use URL
parsing plus explicit component checks; do not fetch or resolve destinations.
Root-only is deliberate: this change does not claim SPA subpath-hosting support.

Build `/t/{path-encoded task ID}` and add `sessionId` with a query encoder. A
backend URL helper must match existing Go `webpush.TaskURL` and TypeScript
`linkToTask` route semantics; parity tests cover spaces, slash, `%`, `?`, `#`,
`&`, `+` and Unicode without double encoding. Use parsed URL composition, not
unescaped concatenation. Verify an actual task/session binding using the event's
pair; if lookup fails, make session identity unavailable, never choose another
session. The notification service needs a narrow session-reader dependency for
this check. It must not change existing recipient/default-copy behavior.

`release.url` accepts only absolute HTTPS without credentials/control characters;
it is distinct from the installation origin. Invalid or absent values are omitted.
An empty configured origin leaves all absolute in-app URL variables absent.
Return `public_url_configured` and safe missing-value diagnostics in the catalog,
never a suggested private/guessed origin. Existing default copy stays unchanged,
including not automatically appending new links.

## Persistent model and API

Add `notification_templates` owned by the existing notification store:

- Composite primary key `(provider_id, event_type)`; `user_id` retained for scoped
  predicates, sourced from the owned provider, never accepted from JSON.
- `template_json TEXT NOT NULL`: serialized `{version:1,title:string,body:string}`.
- `revision` opaque random UUID token replaced on every successful write,
  `created_at`, `updated_at` UTC timestamps. Never reuse a deleted row's revision;
  a stale writer must not match a deleted-and-recreated override (ABA).
- Provider foreign key with delete cascade, plus owner-checked SQL for all reads
  and writes. An INSERT SELECT from the owner/type-filtered provider prevents
  TOCTOU/orphan insertion. Update/delete recheck owner/type inside the transaction.

Proposed authenticated REST endpoints (all under `/api/v1`):

| Method/path | Contract |
| --- | --- |
| `GET /notification-providers/:id/templates` | Owned Apprise provider's catalog and six event records; each has defaults/examples, variable metadata, nullable override and revision. Never include provider config/receiver URLs. |
| `PUT /notification-providers/:id/templates/:event` | `{version:1,title,body,expected_revision}`. Explicit null revision creates only if absent; a saved opaque token updates only that version. Returns canonical override and fresh revision, 200. |
| `DELETE /notification-providers/:id/templates/:event` | `{expected_revision}` body; version check for an existing row, already absent succeeds (204). Restores dynamic defaults, not a saved snapshot of old copy. |
| `POST /notification-providers/:id/templates/:event/preview` | `{version:1,title,body,sample}`; sample is a closed event-specific synthetic scenario key. Returns `{title,body,used_fallback,diagnostics}` (200) without persisting. |

Provider ownership is checked before revealing catalog/event/type details. Foreign
and missing providers return the same 404. Owned non-Apprise provider, unknown
event, malformed/unknown JSON keys, unsupported versions and invalid templates
return 400 with stable error codes and field/key offsets. Bound JSON request
bodies at 128 KiB (413), allowing JSON escape overhead for bounded sources; no
content echo in validation errors. Conflict is 409;
storage failure 500 without raw SQL. No user ID, arbitrary data map, task ID or
session ID is accepted by preview. Responses use `Cache-Control: no-store`.
Controller identity resolution follows the existing synthetic/default-user path.

Separate template operations do not modify provider subscriptions or credential
config. Shared page Save may partially commit independent operations, as today;
track each acknowledged baseline, retain outstanding drafts, and reload after
an uncertain transport outcome before retrying. Revision guards prevent blind
overwrite. No new all-provider atomicity promise.

## Migration, restart and backup/restore

Use additive, replayable `CREATE TABLE IF NOT EXISTS` in the notification schema
owner, with compatible SQL/rebinding for SQLite and PostgreSQL. Existing rows
need no template backfill. Register the table in
`internal/persistence/requiredstores/catalog.go`, conformance coverage and the
explicit-tag upgrade fixture/manifest. Do not edit only fresh-database DDL when
later column evolution is needed. Initialization failure must propagate.

Full database backups contain template rows/JSON/revisions without a separate
export file; restoring an older backup without the table recreates it empty on
upgrade. Restoring a newer backup with the same supported schema preserves rows.
Document backup-before-upgrade, SQLite restore/restart and PostgreSQL pg_dump
procedures. Do not promise arbitrary downgrade compatibility. Invalid future JSON
versions remain stored but unused, with safe fallback and visible diagnostics.
The public URL is operator config and must be backed up separately from the DB.

## Delivery integration

Keep recipient resolution, subscription filtering, delivery claims and runtime
batch membership unchanged. After those checks, build immutable default/context
data, select an owned Apprise override, render and pass a copy of the message to
the adapter. Do not mutate shared payload maps/messages between providers.
Refactor `dispatchGenericNotification` only enough to retain inbox type/title and
workspace destination context and share the Apprise customization seam.
`HandleAgentRuntimeUpdates` must use it after constructing the claimed-member
message, without changing member deduplication or rollback.

Rendering failure is not transport failure: send defaults once under the same
delivery claim. Actual send failure keeps the current claim-release behavior.
No extra send, new queue, cache or external lookups are introduced by templates.
The existing `/test` route bypasses overrides and retains fixed test copy.

## Preview and settings

Backend owns the same catalog/parser/default builder for preview and delivery.
Closed synthetic samples include each event, missing title/session/origin, release,
four single-runtime outcomes, summary and missing conditional fields. Real-data
preview is excluded: it otherwise needs task/session pair authorization and a
separate content-minimization design. Preview has no provider sender dependency
and cannot write deliveries, execute Apprise or contact any receiver.

Extend the existing settings state/hooks/API clients; no direct component fetches.
Add a separate Templates group below the subscription group, selecting an owned
saved Apprise provider and event. New unsaved providers explain that template
editing becomes available after provider creation. Catalog loading must not reset
active drafts. Credentials never appear in this group or its API payloads.

Show read-only subscription state next to Default/Custom status. The template
form has title, body, variable insertion/help with availability notes, current
default samples, preview sample selector, rendered plain text, restore defaults
and field-associated errors. Changes and restore stage into the existing shared
Save/Reset/navigation guard. Record raw submitted drafts/revisions before await;
acknowledgements advance the baseline but do not erase newer edits. Capture a
provider/event/draft request generation for preview; discard stale responses.

The same stacked controls work on desktop and phone, with 44px coarse-pointer
targets, wrapping variable names and bounded textareas. The shared page owns
scrolling. Localize copy in all required catalogs through existing i18n tooling.
Preview text has no HTML or clickable unvalidated links. Mark sample data clearly.

## Security and observability

Use explicit values only: no provider config, receiver URL, token, transcript,
question, permission command, environment or arbitrary payload reflection.
Treat user literals as personal configuration; warn in help not to paste secrets.
No software can reliably recognize every secret voluntarily typed as literal
text, so do not claim generic secret detection. Do not automatically populate
credentials into any source, fixture, preview or diagnostic.

Replace Apprise's raw combined-output error propagation with bounded safe error
classification (missing CLI, timeout, cancellation, nonzero exit). Discard raw
output; never log argv/config/output. Keep the existing 10s cancellation bound,
literal title/body arguments and verify option-looking content cannot change CLI
behavior. Fallback logs may contain provider ID, event key and closed reason code
only, not template text, values, message body, receiver URLs or errors carrying
subprocess output. No high-cardinality metrics are needed.

## Verification and contribution gate

See [plan and work orders](../../../plans/apprise-event-templates/plan.md) for
AC-to-test mapping, exact commands and public-doc delivery. New names in that
plan are proposed test files, not evidence that tests exist or pass.

Before production/test edits, discuss the API/model and missing-value/URL choices
in a feature issue. Do not open even a draft PR while maintainer discussion is
missing (`CONTRIBUTING.md:19-31`). Refresh latest main and reconcile changes
after discussion, then complete the implementation gate and test package. An
eventual fork-to-upstream PR must link the issue and record real results,
screenshots/manual evidence and unexecuted PostgreSQL/platform coverage.
