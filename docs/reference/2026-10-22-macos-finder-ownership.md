# macOS applet ownership repair, 22 October 2026

## Meaning

The Finder review exposed a live-path ownership race and an interrupted-cleanup recovery wedge. The repaired functional cut `94d08bf58d19ee0863e8618d107b29212d1f1778` preserves late edits through a journaled quarantine and accepts already-completed cleanup only after destination reconciliation. It passed the integration regressions, all eleven Verify jobs and both native Finder journeys. This record preserves that evidence separately from the earlier native cut, final product acceptance, physical dragging, signing and publication.

## Repair and preservation boundary

[The complete independent review](https://github.com/Wolfsblvt/smolpng/issues/71#issuecomment-6336245299) found that removal verified live bundles before deleting later objects, and that an absent transaction directory could permanently retain `Pending`. [Nyxara's disposition](https://github.com/Wolfsblvt/smolpng/issues/71#issuecomment-6336531683) accepted both findings for bounded repair while retaining the applet capability and existing engine.

Removal and replacement now journal the intended operation, atomically move the named applet into a private sibling workspace, and verify that moved object. A mismatch restores the object when the destination is absent. A newly occupied destination preserves both objects and refuses. Replacement cannot publish over that occupant. Removal recursively deletes only the verified quarantined bundle. Interrupted removal records its intent; recovery distinguishes an untouched old installation from a completed removal.

Recovery reconciles the visible destination before cleaning up. Missing stage, backup or workspace paths are idempotent cleanup after an exact admitted old, published or removed state. Missing or contradictory destinations are still refused. This closes ordinary named-path races; a hostile same-user process deliberately retaining and mutating open handles inside the private workspace remains outside the selected trust boundary.

## Deterministic evidence

Real temporary filesystem tests exercise four late-edit/replacement cases across installation/removal, two newly occupied destinations, four missing-workspace destination states and four interrupted-removal phases. They also prove repeated repair and subsequent mutation. A narrow manager-local atomic-move seam makes the changes deterministic; actual directory moves, files and ownership digests remain real.

Two temporary negative controls were observed to fail: restoring live-path verification before the move broke both replacement preservation cases; restoring the original strict missing-workspace cleanup broke both old/published recovery cases. These controls were applied to the repaired source, not new tests run unchanged against the original reviewed head. Both controls were removed. The full integration package passed again, as did package vet. After preserving replay onto `main@53ff54e8`, the integration package passed in 3.537 seconds. All ten contribution commits were unchanged according to `git range-diff`; the accepted installer record and qualifier-test fix were preserved.

[Verify run 39751305970](https://github.com/Wolfsblvt/smolpng/actions/runs/39751305970) passed all eleven jobs on exact functional head `94d08bf58d19ee0863e8618d107b29212d1f1778`: three native test/vet/corpus jobs, six cross-builds and two Windows shell builds. Linux used the race detector. Forgehand succeeded; the legacy size job was skipped, not passed.

## Native consumer evidence

[Native Finder run 39751305804](https://github.com/Wolfsblvt/smolpng/actions/runs/39751305804) passed both jobs without retry on the same exact head:

| Host | Image | Application SHA-256 | Native journey |
| --- | --- | --- | --- |
| macOS 15.7.9 / 24G830, amd64 | `macos15` / `20260824.0482.1` | `77868d339a0212a4c986ab0e9062c58ba86c9f1467b62e785e2086872695c935` | Complete, 20.21 s |
| macOS 26.6.2 / 25G83, arm64 | `macos26` / `20261013.0424.1` | `0066e90f31f996d81c0afb504e7d3b76b85489c0ea3e05ad54b12738ecde79a3` | Complete, 21.92 s |

Each journey retained 22 command receipts: twenty successful commands and two expected exit-5 modified-content refusals. Actual native file-open events reached background PNG/JPEG compression and both Terminal variants; originals and independent output decoding were checked. Reinstallation, modified-file preservation and complete removal passed. Both final status receipts contain empty entries and `recoveryPending=false`. Native behavior and filesystem interruption regressions prove their respective boundaries; the native journey does not simulate a process crash during deletion.

The unsigned applet, independent optimizers, generated fixtures and disposable Mac account boundaries remain unchanged. No clipboard, remote quota, Apple credentials, privacy grant or public package was used. The [earlier native record](2026-10-22-macos-finder.md) retains its failed preparations and macOS 26 Intel image limitation; those observations were not overwritten or counted as repaired-host proof.

## Original evidence bytes

[report.json](2026-10-22-macos-finder-ownership/report.json) retains all **80 original files**, **199,246 decoded bytes**, from the two repaired-head native specimens. Each base64 value reconstructs one exact file; length and SHA-256 were verified after serialization. Source/host/tool identities, all 44 command receipts, Terminal transcripts and complete reports are included. Executables and app bundles are excluded.

The payload is **286,032 bytes**, SHA-256 **`4582a65c3047003916336d5669cf368bfefc0a4012fe4b2125cfb5b7d87ad4ea`**. The original hosted artifact digests are `dcab799eef2dd5ec9e5fabefc1a8517d1158313d009590151de935e830c955f2` (Intel) and `06b1b523d698ff35d70979321b9417c0424dacf1289de7e08b7b0203ca15b8b3` (ARM64).

This record carries the repaired functional cut; any later documentation-only commit does not rewrite its source identity. Final-head disposition and remaining product custody stay in [Work #71](https://github.com/Wolfsblvt/smolpng/issues/71). No second independent review round, private joined candidate, release or publication follows merely from these checks.
