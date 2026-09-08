# Security boundaries

## Meaning

smolpng is a trusted-user desktop/CLI tool, not a public upload service. Its current protections cover bounded parsing, direct process arguments and original-preserving output. No independent audit, public security intake or response-time promise is claimed.

External optimizers are trusted executables running as the invoking user. They receive temporary copies, not original paths, and no shell is used. Local optimizer environments exclude TINIFY_API_KEY. Raw exec intentionally relinquishes managed-output guarantees. PNG CRC/dimensions and input bytes are checked; animated compression is refused. URL input is explicit, bounded and redacted, but internal URLs are allowed for a trusted caller.

Outputs are atomically visible, not claimed power-loss durable. No-clobber writes refuse on unsupported filesystems rather than falling back to partial copy. In-place rechecks bytes but is not an adversarial filesystem lock. Process cancellation kills the directly owned child; arbitrary descendants are not separately supervised. Real desktop, keyring, managed download/signing and web ingress have their own future qualification boundaries.

## Skill management

Only explicit dedicated `smolpng` directories are tracked. Resource hashes gate synchronization and uninstall; symlinks, untracked bytes and unexpected empty directories are preserved. A journal and native directory moves recover interrupted updates without dropping unknown content. This is not an adversarial filesystem sandbox: a malicious concurrent owner of the containing directory remains outside the model. Corrupt registries and ambiguous recovery stop with diagnostics. App-managed and custom-path update consent are distinct.
