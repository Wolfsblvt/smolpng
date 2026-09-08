# ChatGPT Skill and executable boundary

## Meaning

The supported Skill package contains instructions and references, not an embedded desktop executable. A native program successfully running in one ChatGPT authoring sandbox does not prove that the Skill uploader, scanner, install surface and future runtime will admit or execute that program. Keep those observations separate.

## Current published evidence, checked 2026-09-08

OpenAI's **MCP-to-plugin Skill importer** documents 256KiB for SKILL.md, 1MiB for each supporting resource, 5MiB for all resources in one Skill, and 8MiB for generated scan archives including ZIP overhead. These are importer limits, not a blanket statement that every manual Skill-upload flow has identical limits. Its resource digest is over the decoded file bytes; compressing a large executable inside the upload ZIP does not make that executable satisfy a per-resource uncompressed limit.

The Help Center separately describes personal Skill upload with supporting files/code and a scan that can mark the Skill as needing review or blocked. The page reviewed does **not** publish the same numerical per-file limit for that manual flow or guarantee arbitrary native executable support. An actual manual-upload/install/run experiment would still be required; documentation should not invent either acceptance or rejection on that surface.

Primary sources:

- https://developers.openai.com/plugins/build/mcp-server#import-skills-from-the-mcp-server
- https://developers.openai.com/plugins/build/skills
- https://help.openai.com/en/articles/20001066

## Three independent requirements

**Artifact size:** measure the actual candidate executable as well as its ZIP, not the whole repository. Stripping debug/symbol data may shrink a binary, but the measured result still needs to fit the chosen upload/import contract. Do not split or wrap resources to evade a limit. The instruction-only Skill remains a much smaller separate artifact.

**Execution environment:** Windows PE executables are not Linux ELF executables. The authoring container can be inspected with ordinary system tools and tested with the matching Linux build. That does not establish a permanent ChatGPT sandbox OS/architecture contract or native-file scanner approval. Browser chat has no implied access to an executable on the user's desktop.

**Optimizer dependencies:** the standard application deliberately does not bundle pngquant, OxiPNG or jpegoptim. A single smolpng binary provides orchestration, validation and some explicit in-process preparations, not all ordinary local compression backends. Even a successful binary upload still needs those separately licensed tools in the runtime, or an explicitly configured/authorized remote provider. Do not silently download or upload merely to make a Skill demo work.

## V1 posture and a bounded experiment

Keep the v1 Skill instruction-only for local agent harnesses and a future authenticated hosted app. An optional experiment can attach the correct platform candidate separately, verify its checksum, run `version --json` and `capabilities --json`, then compress one project-owned specimen in a temporary directory. Record platform, source SHA, exact byte sizes, output and every untested step. A no-optimizer already-fitting result or explicit matte conversion proves binary execution but not the full optimizer pipeline.

If testing the actual manual Skill uploader later, use its normal controls, preserve the original packaging and scanner result, and do not mislabel files or work around resource limits. A successful test is evidence for that exact surface/date, not a new universal installation guarantee. The future smolpng-web authenticated app remains a separate hosted workflow; no local MCP server is added by this experiment.
