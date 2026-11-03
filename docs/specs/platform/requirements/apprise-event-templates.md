---
status: draft
system: platform
created: 2026-11-03
owners:
  - kandev
---

# Apprise Event Templates Requirements

## Overview

Users need to manage the title and body sent by each Apprise provider without
patching an installation-specific domain or notification string. Platform owns
this vertical capability because it owns notification preferences, delivery and
configuration. This proposal extends, rather than replaces, the existing
[semantic notification contract](notifications.md).

Upstream proposal: [#5011](https://github.com/kdlbs/kandev/issues/5011).
The user subsequently directed delivery to the Illuwa fork and deployment on
the existing Mac runtime. The runtime source is preserved as `production/mac`
at `2d8aee5817645b4e8518e56d5d268f0addac1c79`. This compatibility delivery
covers its four offered events; it does not add newer upstream events or
upgrade unrelated runtime behavior. Upstream contribution remains a separate
discussion gate. The fork's issues are disabled; the explicit owner request
and this design package record the fork-local delivery decision.

## Terminology

- **Override:** one custom title/body pair for one provider and event key.
- **Default:** the current backend-generated title/body pair for that occurrence,
  including its existing missing-task and runtime-update variants.
- **Preview:** rendering against labelled synthetic data without delivery.
- **Destination:** a navigable URL appropriate to the actual event context.

## Requirements

### REQ-PLATFORM-APPRISE-TEMPLATES-001: Owned event template management

**Intent:** Let users control Apprise message content independently of delivery policy.

#### Acceptance criteria

- **AC-PLATFORM-APPRISE-TEMPLATES-001.1:** A user shall create, read and update one
  title/body override for each of their Apprise providers' four supported event
  keys: `session.turn_finished`, `session.clarification_requested`,
  `office.inbox_item`, and
  `system.update_available`. Deleting an override shall restore the current
  Kandev default for subsequent occurrences; repeated deletion shall succeed.
- **AC-PLATFORM-APPRISE-TEMPLATES-001.2:** Templates and event subscriptions shall
  be separate preferences. Creating, saving, previewing or deleting an override
  shall not subscribe, enable or send through a provider. An unsubscribed event
  or disabled provider shall not send even when an override exists.
- **AC-PLATFORM-APPRISE-TEMPLATES-001.3:** CRUD and preview shall be limited to the
  provider owner. A foreign provider shall be indistinguishable from a missing
  provider. Changing a provider away from Apprise shall remove its overrides
  atomically; changing it back shall not resurrect them.
- **AC-PLATFORM-APPRISE-TEMPLATES-001.4:** Without an override, all existing event
  copy, recipients, subscription defaults, deduplication and retry behavior shall
  remain unchanged. Local, System, Web Push and Webhook shall not use Apprise
  overrides. The existing explicitly sending provider-test action shall remain
  distinct from preview and retain its existing fixed test message.

### REQ-PLATFORM-APPRISE-TEMPLATES-002: Bounded event variables and destinations

**Intent:** Expose truthful context and links without executing user templates.

#### Acceptance criteria

- **AC-PLATFORM-APPRISE-TEMPLATES-002.1:** Each event shall publish a variable
  catalog with descriptions, availability conditions and synthetic examples.
  Unknown variables, variables belonging only to another event, malformed
  placeholders, functions, loops and expressions shall be rejected before save
  and preview. Values shall be substituted once as literal text, never reparsed.
- **AC-PLATFORM-APPRISE-TEMPLATES-002.2:** Only session occurrences carrying an
  actual task/session pair shall expose `session.id` and `session.url`. A missing
  session or a pair that cannot be verified shall never be replaced with a
  primary, latest, sibling or invented session. Missing referenced values shall
  cause the entire title/body pair to fall back to defaults with a safe diagnostic.
- **AC-PLATFORM-APPRISE-TEMPLATES-002.3:** Operators shall configure an optional
  install-wide public web origin. Generated links shall use HTTPS except for
  HTTP loopback development. Credentials, query, fragment, non-root path,
  unsupported schemes and malformed ports/hosts shall be rejected. Unconfigured
  origins shall leave absolute in-app URLs unavailable, not guess a domain from
  request headers. IDs shall be separately path/query encoded. No deployment
  domain shall be built into the product.
- **AC-PLATFORM-APPRISE-TEMPLATES-002.4:** System updates shall expose only
  applicable update destinations, with no task or session variables. Release,
  single-runtime and runtime-summary shapes shall be documented and previewable.
  Office inbox notifications shall link to the inbox without inventing an item,
  task or session identity. Invalid external release URLs shall be unavailable.
- **AC-PLATFORM-APPRISE-TEMPLATES-002.5:** Title sources shall be at most 1,024
  UTF-8 bytes, body sources 16,384 bytes, and each field at most 64 placeholders.
  Rendered titles shall be at most 2,048 bytes and bodies 65,536 bytes. Invalid
  UTF-8, NUL, and multiline titles shall be rejected. Oversized or failed custom
  rendering shall fall back as a complete pair. Rendering shall have no code,
  file, environment, network, recursion or unbounded expansion capability.

### REQ-PLATFORM-APPRISE-TEMPLATES-003: Safe settings editing and preview

**Intent:** Make customization discoverable and recoverable on desktop and phone.

#### Acceptance criteria

- **AC-PLATFORM-APPRISE-TEMPLATES-003.1:** Notifications settings shall show an
  Apprise provider/event selector, separate subscription status, default/custom
  status, labelled title/body controls, event-specific variables, default copy,
  restore-default action and preview. Restore shall stage deletion until the
  shared Save succeeds. Shared Reset shall discard drafts. Empty source fields
  shall not be interpreted as deletion and shall be rejected.
- **AC-PLATFORM-APPRISE-TEMPLATES-003.2:** Preview shall use labelled synthetic
  examples only, permit choosing documented event shapes/missing-value examples,
  and display rendered text, missing-variable warnings and fallback state. It
  shall neither read real tasks/transcripts nor invoke any provider, network
  delivery or delivery-history write. Results shall render as text, not HTML.
- **AC-PLATFORM-APPRISE-TEMPLATES-003.3:** Save/preview errors shall be visible and
  accessible without losing drafts. Successful acknowledgements shall advance
  only the submitted baseline, preserving later edits. Stale preview responses
  shall not overwrite another provider/event/draft. Concurrent saved edits shall
  report a conflict requiring reload rather than silently overwrite.
- **AC-PLATFORM-APPRISE-TEMPLATES-003.4:** Desktop and phone shall expose the same
  operations without horizontal page scrolling. Phone/coarse-pointer actions
  shall have at least 44px targets. Keyboard operation, field-associated errors,
  loading/status announcements and all supported UI translations shall be covered.

### REQ-PLATFORM-APPRISE-TEMPLATES-004: Durable and confidential configuration

**Intent:** Preserve configuration across restart and upgrades without leaking credentials.

#### Acceptance criteria

- **AC-PLATFORM-APPRISE-TEMPLATES-004.1:** Saved overrides shall survive backend
  restart and full database backup/restore. Upgrade and repeated initialization
  shall preserve existing providers, credentials, subscriptions and delivery
  history, create no implicit overrides and never re-enable removed subscriptions.
  Provider deletion shall remove its overrides. Confirmed write failures shall
  leave the prior template intact.
- **AC-PLATFORM-APPRISE-TEMPLATES-004.2:** Template catalogs, previews, errors and
  logs shall never include provider receiver URLs, authentication tokens or
  subprocess output containing those values. No credential/config object shall
  be available as a template variable. No rendered content or template source
  shall be recorded by new diagnostics.
- **AC-PLATFORM-APPRISE-TEMPLATES-004.3:** Public documentation shall explain
  setup, syntax, event variables and missing values, defaults, URL configuration,
  API permissions and errors, upgrade and backup/restore behavior. Submission
  shall report actual unit/integration/E2E/manual results and any unexecuted
  coverage; it shall not claim implementation readiness from document checks.

## Out of scope

- New event subscriptions, transport types, quiet hours, retry queues or HTML mail.
- Template execution, user-provided variable resolvers or real-data previews.
- Templates for Local, System, Web Push, Webhook or plugins.
- Production deployment or changes to any live instance's configuration.

## Related design and delivery

- [System design](../system-design/apprise-event-templates.md)
- [Implementation plan](../../../plans/apprise-event-templates/plan.md)
