# Mac Runtime Provenance and Rollout

## Baseline preservation

On 2026-11-03 the installed health endpoint reported
`0.98.0+upstream-automation-history.1`. The existing fork PR #4 is merged into
`compat/v0.98.0` at `2d8aee5817645b4e8518e56d5d268f0addac1c79`.
Its source commit and binary SHA-256 match the retained deployment receipt:
`17b6c0726bc4c563e0a6a5a140b0fb81fcb250d5f2008f18a602219051e9bb02`.
The same commit was pushed without rewriting history to `production/mac` in
`illuwa-soft/kandev`. Main was not reset or replaced. The feature PR targets
`production/mac`, so the reviewed diff contains only this capability.

The existing Hindsight automation-history contract is already in that base.
No plugin archives, credentials, instance IDs or workspace bindings are changed
by the feature. Calendar session-link visibility is a related use case, verified
by the merged [Stalwart PR #2](https://github.com/illuwa-soft/kandev-plugin-stalwart/pull/2),
not evidence that notification-template tests or deployment already passed.

## Rollout procedure

1. Complete scoped test/type/lint/schema and desktop/phone browser checks.
2. Build a production (not E2E/mock) web bundle and native backend with the exact
   feature commit and distinct version linker values. Retain the old bundle.
3. Back up the database with SQLite's native backup API, independently restore
   to a private temporary file, and compare integrity and logical data. Back up
   operator configuration/launcher and hash plugin metadata without printing
   provider configuration or secrets.
4. Install the candidate as a separate private bundle. Change only the existing
   launcher target and the operator public-origin setting. Never replace the
   manually managed LaunchDaemon or alter plugin policies/bindings.
5. Restart via the existing service manager. Because the current task runs in
   Kandev, run the bounded transition/checker independently of the service and
   persist a private receipt. On startup failure, restore the original launcher
   and binary; do not rewind the live database or external plugin history.
6. Read back version/readiness, template catalog/preview using an owned disposable
   disabled provider, and delete that provider afterward. Check that existing
   provider/config/plugin identities are preserved using hashes/counts only.
   No test emails or receiver calls are required for this smoke test.

## Results

Baseline preservation completed. Feature PR, production artifact, backup drill,
transition and post-restart smoke: pending. This is not a deployment receipt.
