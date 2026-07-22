# wagie-templates

## Scope

This repo is a companion library of specialized Wagie templates that do not belong in
Wagie's core root-level library.

Families are shallow root-level directories, one per domain. Place templates by the
domain that owns their meaning, not by incidental tooling. Do not reintroduce a
`templates/` wrapper, and do not create deep category trees unless the repo grows
enough to justify them.

## Building Blocks

A template is one responsibility with a declared boundary: typed inputs and outputs,
composable via `uses:` or usable standalone.

A seam earns its own template only when it is real: reused by more than one plausible
caller, retrievable as a concept a searcher would query, or carrying genuine
orchestration (a loop, fan-out, distinct worker placement, or a typed gate) behind a
clean single-responsibility boundary. Otherwise fold the work into the parent's
`instructions` or a container step — prefer a little duplication over the wrong split.

Fragment for orchestration leverage, not to hand-hold the agent. Wagie's value here is
loops, fan-out, worker placement, and enforced typed handoffs between steps — the things a
single worker can't do in one pass. It is *not* chopping a worker's reasoning into many
micro-tasks. Give a worker a whole cohesive job and a typed output, and let external
knowledge surfaces (domain runbooks, docs) carry the domain procedure. A step that is
just "the agent thinks" belongs inside a worker, not as its own task — and every extra
task also dilutes the retrieval pool the router searches.

## Family Shape

Templates can be user-facing entry points, pipeline stages, or internal building
blocks. Do not maintain template-name inventories in this file; they drift.
Use each template's `description`, `tags`, inputs, and outputs to signal whether
it is an entry point or a building block.

- `ethereum/`: keep Ethereum devnet and Kurtosis workflows pipeline-shaped.
  Templates should be thin orchestration/glue around Panda runbooks: gather or
  generate config, provision or snapshot a network, observe it, then investigate
  structured issues. Panda runbooks own domain procedure and operational
  knowledge; templates own typed handoffs, fanout, loops, and artifacts.
- `code/`: keep code-review, source-investigation, verification, and fix-loop
  workflows here when they are code-domain specific. Generic review-loop or
  consensus primitives belong in core Wagie.
- `research/`: keep iterative research workflows here when they coordinate
  search, findings, coverage assessment, and verification around a research
  question. Generic extract/summarize/evaluate primitives belong in core Wagie.
- `experiments/`: keep metric-driven experiment loops here. The common shape is
  discovery/setup, propose/apply/check/measure, keep-or-revert, then summarize
  progress and convergence.

When authoring new templates, flag the role in the template description if it is
not a top-level entry point.

## Boundary

Templates in this repo should be specialized, domain-coupled, or operator-specific.

Keep generic material in core Wagie: atomic cognitive primitives, generic
orchestration patterns (review loops, consensus, map-reduce), and structural glue.
If a template can stand as a root-level composable primitive for many unrelated
workflows, it probably belongs in core, not here.

## Repo Rules

- Prefer targeted validation while iterating, then run the smallest meaningful broader check before finishing.
- Keep instruction files concise and scoped; put detailed topic guidance in referenced docs instead of growing this file.
- Prefer the smallest family that owns the workflow. Avoid dumping unrelated templates into a catch-all bucket.
- Cross-family dependencies should be rare and intentional. Prefer depending on core Wagie templates over coupling families together without a strong reason.
- References to external knowledge surfaces (e.g. runbook refs inside instructions) are
  plain prose to the worker — nothing here validates them, so keep them current when the
  external surface changes.

## Core Commands

```bash
make test                       # full combined-library validation (go test ./...)
make validate                   # per-file ok/FAIL report across all families
make validate FILTER=code       # report only paths containing "code"
make tidy
```

`make test` and `make validate` both validate domain templates against core
templates from wagie's embedded Go module, via wagie's executable preflight
(parse + compile), against the pinned `github.com/savid/wagie` release in
`go.mod`.

The strict parser rejects unknown fields and fails one error per file at a time;
re-run after each fix until `0 failed`. See `docs/agent-guides/project-templates.md`
for the spec grammar (planes, leaf vs container outputs, matrix/loop, secrets).

## Template Work

When editing `**/*.yaml`, also read:

- `docs/agent-guides/project-templates.md`

Template authoring is retrieval-sensitive. Names, descriptions, tags, inputs, `uses:`, task names, and output contracts all affect how Wagie finds and composes templates.
