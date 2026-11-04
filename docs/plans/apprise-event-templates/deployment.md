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

Completed and independently read back on 2026-11-03 at 10:08:40 UTC:

- Feature PR: [Illuwa #5](https://github.com/illuwa-soft/kandev/pull/5), open at
  rollout time. Runtime source: `15495808e383c0af926a617b54cd8b2adacd9db3`.
- Running version: `0.98.0+upstream-history.apprise-templates.1`, verified through
  local health, readiness and the existing public endpoint.
- Installed native SHA-256:
  `29e8d41280b010a918511a7114586695416bae630fd804e98fe4cb84c3b4ebb6`.
- Production Node 24.15.0 web bundle embedded in a Go 1.26.0/fts5 native build,
  with explicit version/commit/time and `-buildvcs=false`. Existing helper
  binaries retained; no plugin or automation-history code was replaced.
- The exact candidate binary also passed isolated managed Chromium and Pixel-5
  scenarios, one each. Reviewed synthetic viewport captures are in
  `docs/screenshots/apprise-templates-{desktop,mobile}.png`.
- Private SQLite native backup and independent restored copy passed integrity
  checks and exact snapshot/restored-file hash comparison. The initial live
  copy timed out before operational mutation; pinning a WAL read snapshot
  allowed a bounded consistent copy while background writes continued.
- A separate session-independent worker changed only the existing launcher's
  bundle/version target and public-origin environment value, then signalled
  the verified launchd-owned PID. KeepAlive restarted the existing service.
- An owned temporary disabled provider proved catalog availability, template
  create/read/update/delete, synthetic preview and missing-session whole-pair
  fallback. The temporary provider was deleted. Test sends: **0**.
- Existing provider configuration, plugin installation/version/status inventory,
  feature flags, key/configuration files and service definition matched their
  pre-transition hashes. No live database rollback or plugin-history rewind.
- The old bundle and launcher are retained. Private receipt and restore artifacts:
  `~/.kandev/backups/apprise-templates-20261103/`. No credentials or raw receiver
  configuration are included in this public record.

`deployed/mac` points to the exact runtime source. At PR creation, `production/mac`
retained the pre-feature baseline; merging PR #5 advances that integration branch.
The old source remains identified by the baseline commit above. Follow-up
screenshot/test/receipt changes, Go 1.26 test formatting and an equivalent
`available` string-constant extraction do not install a new binary. Full CI and
real PostgreSQL completion evidence is recorded in
[the plan](plan.md#full-validation-completion). Upstream proposal
[#5011](https://github.com/kdlbs/kandev/issues/5011) receives this fork validation
as evidence, not as a claim of upstream acceptance or six-event delivery.
