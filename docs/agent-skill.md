# Agent Skill and installation lifecycle

## Meaning

The product Skill teaches agents how to use an installed smolpng; root `AGENTS.md` teaches contributors how to develop it. The Skill and references are embedded in the executable from `skills/smolpng/`. Its version is a content hash, not the application version, so a release with unchanged instructions does not create needless synchronization prompts.

Local coding agents must have a host able to execute smolpng. Packaging or uploading this Skill does not let ChatGPT Web execute a user's desktop binary. No executable, remote endpoint, credential, or local MCP server is included.

## Inspect and install

```sh
smolpng skill status --json
smolpng skill install --target agents --json
smolpng skill install --target agents --apply --json
```

Install previews are read-only. `--apply` writes the displayed dedicated `smolpng` directory, with every file's SHA-256 recorded in `skill-installations.json` under the application state directory. Existing untracked directories are never adopted or overwritten, even when their content happens to match.

Targets:

| Selection | Destination |
| --- | --- |
| `auto`, `agents`, `codex` | `~/.agents/skills/smolpng` |
| `claude` | `~/.claude/skills/smolpng` |
| `project` | `<current-directory>/.agents/skills/smolpng` |
| `--path <parent>` or `--target path:<parent>` | `<parent>/smolpng`, marked custom |

Auto selects only the shared personal target. Detection reports observed harness folders; it does not authorize installation into every discovered harness. Custom `CLAUDE_CONFIG_DIR` or other nonstandard layouts use an explicit parent. Project selection means the invocation directory, not an inferred ancestor. Current path evidence is recorded below; harness discovery is not a guarantee that a particular running client has reloaded its Skills.

## State and synchronization

`skill status --json` and `doctor --json` expose exact destinations, hashes, application and bundle versions, whether a destination is standard/app-managed or custom, and:

- `current`: bytes still match their recorded manifest and current bundle;
- `stale`: bytes match their previous manifest, but the embedded bundle changed;
- `modified`: resources were edited, removed or added;
- `missing`: the whole tracked directory disappeared;
- `unknown`: a path is unreadable, contains symlinks/unexpected empty directories, or exceeds inspection bounds.

Only an unmodified stale installation is automatically eligible for replacement. Commands default to preview:

```sh
smolpng skill sync --json
smolpng skill sync --apply --json
smolpng skill sync --apply --include-custom --json
smolpng skill sync --destination /exact/tracked/smolpng --apply --json
```

`update` aliases `sync`. The first applied form updates eligible app-managed destinations only. `--include-custom` or a specific `--destination` supplies custom-path consent. Modified, missing and unknown content is preserved, not silently repaired or replaced. Restore it deliberately, or stop tracking it and install elsewhere. The library's `SyncAfterUpdate` entry point applies the same managed-only rule; the implemented standalone updater invokes it only after signature/hash and new-build identity verification.

```sh
smolpng skill uninstall --destination /exact/tracked/smolpng --apply
smolpng skill forget --destination /exact/tracked/smolpng --apply
```

Uninstall removes only an unchanged tracked directory, or unregisters a wholly missing one. Forget changes tracking only and leaves all content intact. Neither operation has an overwrite-modified shortcut.

## Executable replacement

Installation/update records the last-seen application and embedded bundle identity. Read-only status detects a later executable version or bundle change. Only changed Skill content with outstanding installations earns a synchronization prompt; an application version bump alone does not. An interactive-startup observer retains a pending condition until handled, rather than consuming it before a human sees it. `skill ignore-update` remembers the current bundle hash; a different future bundle can prompt again.

The predicate, persistence and keyboard startup observer are implemented and tested. The human sees synchronization, review, later and ignore-current-bundle choices only when the embedded content changed and a tracked destination is outstanding. Current deterministic commands do not prompt or contaminate stdout, and diagnostic reads do not create a first-run state file.

## Interrupted updates and filesystem safety

A native OS advisory lock serializes app management operations; it is released if the process dies. A private sibling directory is fully staged and hash-verified before directory swaps. A journal records the previous and next installation, staging path and backup before any destination move. Atomic no-replace moves prevent an unexpected existing destination from being silently replaced.

After interruption, a subsequent write or `skill repair --apply` completes an already-present exact new bundle or restores the exact old bundle. Unexpected user bytes stop recovery. Backups are deleted only when their complete manifest still matches; changed leftovers remain listed as pending cleanup. Read-only status reports a pending transaction without altering it.

The installer refuses symlink paths and cannot manage arbitrary directories: each destination must be a dedicated `smolpng` child. This is protection for ordinary user-owned storage, not a sandbox against a malicious concurrent directory owner. A filesystem lacking the required atomic directory primitive refuses rather than falling back to partial copying. Whole-machine power-loss durability and real Windows/macOS locked-directory behavior need platform qualification.

## ChatGPT package

```sh
smolpng skill print
smolpng skill package --target chatgpt --output smolpng-skill.zip --json
```

The ZIP contains `smolpng/SKILL.md`, references and MIT license. Entry order, timestamp and hashes are deterministic. Packaging never replaces an existing file. The bundle stays within the documented resource-size envelope, but actual upload scanning/acceptance and client behavior have not been tested here. A future authenticated `smolpng-web` app remains the route for remote execution, not a claim made by this instruction package.

## Primary path/format evidence

Read on 2026-09-08. Recheck when installing into a new harness layout:

- OpenAI, [Build Skills](https://learn.chatgpt.com/docs/build-skills): shared personal `~/.agents/skills` and repository `.agents/skills` discovery.
- Anthropic, [Agent Skills](https://code.claude.com/docs/en/skills): personal `~/.claude/skills` and project `.claude/skills`.
- GitHub, [About Agent Skills](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills): shared `.agents/skills` among supported layouts.
- [Agent Skills specification](https://agentskills.io/specification): `SKILL.md` entry point and frontmatter.

The shipped instruction text and path adapters were checked against those source contracts. No hosted app, authentication, paid provider request or public Skill upload occurred.
