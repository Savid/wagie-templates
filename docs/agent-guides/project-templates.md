# Wagie Templates Guide

Use this guide when authoring or reviewing YAML templates in `wagie-templates`.

This repo is not the core Wagie template library. It is the home for specialized families that are better loaded as an external companion library.

## Library Boundary

Keep templates here when they are:

- domain-specific, such as Ethereum devnet and Kurtosis workflows
- tool-specific, such as code-review automation
- operator-specific or opinionated enough that they should not shape Wagie core
- tightly coupled to one family's concepts, inputs, or operational surfaces

Keep templates in Wagie core when they are:

- generic atomic capabilities like classify, extract, summarize, transform, evaluate, or answer
- generic orchestration patterns like review loops, consensus, routing, map-reduce, or promotion logic
- reusable building blocks that would improve many unrelated domains

If a template would make sense as a core root-level composable primitive, it probably does not belong here.

## Family Placement

The repo uses shallow root-level families:

- `ethereum/`
- `code/`
- `research/`
- `experiments/`

Place templates by the domain that owns their meaning, not by incidental tooling.

Examples:

- a devnet-to-Kurtosis repro workflow still belongs in `ethereum/`
- a code-review pipeline belongs in `code/`
- iterative finding accumulation belongs in `research/`
- metric-driven artifact improvement loops belong in `experiments/`

Do not add an extra `templates/` wrapper. Do not create deep category trees unless the repo grows enough to justify them.

## What Good Looks Like

A good template in this repo is:

- valid under the Wagie spec and consumer validation path
- clearly specialized enough to justify living outside core
- easy for Wagie to retrieve and compose
- smaller because of composition, not larger
- explicit about inputs, outputs, and control flow

## Start From Nearby Templates

Before writing a template:

1. Inspect 2-4 nearby templates in the same family.
2. Check whether the workflow really belongs in this repo rather than Wagie core.
3. Prefer updating a nearby template or reusing a building block over creating an overlapping workflow.

Signal the template's role through naming, description, and composition rather than taxonomy tags. Entry-point templates should have clear descriptions; building-block templates should say so in the description ("Usually called via …").

## Compose First

Default to `uses:` when an existing template already solves part of the problem.

Prefer:

- `uses:` for stable reusable blocks
- small typed data flow between tasks
- thin wrappers around well-named building blocks
- depending on Wagie core for generic primitives instead of copying them into this repo

Avoid:

- duplicating a core template here with family-specific naming
- large opaque `run` tasks that hide orchestration semantics
- adding extra review or evaluation steps that the combined library already has

## Write For Retrieval

Retrieval quality depends on more than tags: name, description, tags, inputs, `uses:`, task names, instructions, and outputs all matter.

### Descriptions

- Start with `Use when ...`
- Keep it to one or two tight sentences
- Include:
  - trigger or situation
  - desired outcome
  - important constraint, tool, or workflow shape when material

Good:

`Use when an active Ethereum devnet issue must be investigated and reproduced in Kurtosis with config choices grounded in the source devnet profile.`

Weak:

`Workflow for Ethereum debugging.`

### Tags

Tags are retrieval search terms, not taxonomy labels. Use the bare keywords a user would type when looking for a template.

Prefer:

- short nouns or noun phrases tied to the template's purpose (`bug-hunt`, `evaluation`, `coverage`)
- family or integration terms when they narrow retrieval (`kurtosis`, `ethereum`, `ethpandaops`, `devnet`)
- accuracy over volume

Do not:

- use `type:*`, `flow:*`, or `cap:*` prefixes
- pile on synonyms
- invent pseudo-taxonomies that the search path doesn't know how to rank

## Design The Contract Before The Tasks

Define:

1. what the caller provides
2. what the template guarantees on success
3. which outputs downstream templates will consume

Every `inputs:`/`outputs:` boundary is split into three planes — `values:`
(JSON data, each carrying a JSON Schema), `artifacts:` (blob-like resources
passed by reference), and `secrets:` (secret references). There is no flat
`inputs: {name: ...}` map.

Input rules:

- each input value lives under `inputs.values.<name>` and carries a `schema:`;
  the JSON type, `default`, `enum`, `properties`, and `items` live INSIDE
  `schema`, while `required:` and `description:` are siblings
- a value cannot be both `required: true` and carry a schema `default:`; a real
  default is materialized at the read site with `.orValue(<default>)`, not by
  `default:` (which is documentation only)
- blob inputs go under `inputs.artifacts.<name>` and secrets under
  `inputs.secrets.<name>`; both are wired by `ref:` and are never read inside a
  `${{ }}` value expression
- keep caller-facing inputs simpler than internal task wiring

Output rules:

- every public output should be intentionally consumable
- a leaf task's `outputs.values.<name>` are PRODUCED by the worker — declare
  `{ required?, schema }` only, with no `value:` expression
- a container/template's `outputs.values.<name>` AGGREGATE child outputs —
  declare `{ required?, schema, value: "${{ ... }}" }`; these expressions may
  read `tasks.*`, `inputs.values.*`, `matrix.*`, `loop.*`, but not the
  `outputs` namespace
- files a worker writes are declared under `outputs.artifacts.<name>`
- keep output keys stable and descriptive

## Task Authoring

A leaf task is identified by having `instructions:` (no child `tasks:` or
`uses:`). It dispatches worker work. There is no `run:`, `selection:`, or
`timeout:` — worker routing is router policy, not authored into the spec.

For leaf tasks:

- bind each input under `inputs.values.<name>: { schema, value }`; declare each
  produced output under `outputs.values.<name>: { required?, schema }`
- add `quality-gate` when malformed or empty output would poison downstream
  steps; it is a CEL boolean over the task's own produced outputs
  (`outputs.values.*`). Any input a gate reads must be `required: true` on that
  bound input
- `retryable` defaults to true for leaves; set `retryable: false` to opt out.
  Attempts, backoff, and timeouts are router-owned
- if a worker capability is genuinely load-bearing (e.g. tool-use), state it in
  one short sentence of prose rather than a routing field
- write instructions that constrain the worker to the exact contract you need

### Prompt Hygiene

Wagie already injects task inputs and output expectations into worker context. Do not restate that machinery in the prompt unless a task has a truly unusual requirement.

Prefer:

- referring to fields by their natural names, such as `finding.description` or `repo`
- concise task instructions focused on reasoning, policy, or transformation logic
- relying on the declared `outputs.values` schemas for the response contract

Avoid:

- dumping input values into the prompt with direct `${{ inputs.* }}`
  interpolation — `${{ }}` is NOT expanded inside `instructions:`; the engine
  passes the string verbatim, so wire values through `inputs.values` and refer
  to them by name in the prose
- sections that only restate already-injected inputs
- repeating output format or JSON shape requirements already declared in the
  `outputs.values` schemas

The schema is the source of truth for outputs. Keep prompts aligned with it, but do not duplicate it.

### Agent-Agnostic Instructions

Templates run on isolated workers. Write instructions for a generic worker, not for a specific product UI or agent brand.

Prefer:

- treating declared inputs, mapped task outputs, and declared artifacts as the complete task context
- saying which side effects are allowed before procedural steps, especially for `gh`, git, Kurtosis, Docker, or networked tools
- using positive, verifiable instructions: what to inspect, what to emit, and what evidence changes the decision
- separating facts, inferences, decisions, and limitations in investigation prompts
- asking for evidence-grounded conclusions rather than chain-of-thought or hidden reasoning
- using examples only as examples; make clear they are not exhaustive routing rules

Avoid:

- product-specific tool names or UI modes inside reusable templates
- assuming shared memory between tasks beyond explicit inputs, outputs, artifacts, or durable external state
- embedding one past incident as a permanent rule when a generic discovery instruction would catch the same class of failure
- long walls of `MUST`/`Never` rules for reasoning behavior; reserve hard language for safety, side effects, and output invariants
- comments about legacy output shapes after a hard cutover

For composed (`uses:`) tasks:

- a `uses:` binding is binding-only: a value input carries exactly `value`, an
  artifact/secret input carries exactly `ref`. `schema`, `required`, and
  `description` belong to the target template — putting any of them on the
  binding is a validation error. Omit a binding entirely to take the target's
  default.
- the bound value still must match the target's declared schema (the sampler
  checks it), so objects/arrays you feed in must match the target's shape;
  reshape with `.map(...)` or use `boundary.assert(...)` when the producer is
  loosely typed
- read a composed task's outputs as `tasks.<task>.outputs.values.<name>`
- list every task an expression reads in `needs:`
- preserve established field names when possible

For loops and matrices:

- `matrix.axes.<axis>: { schema, value }` fans out in parallel; the axis value
  is the array to iterate and `matrix.<axis>` is each element. There is no
  `range()` — iterate the array of objects directly and index sibling arrays by
  a field carried on the element. Reading a child output at the matrix boundary
  aggregates it into a collection across combinations
- `loop: { max, until, onMaxExhausted }` repeats sequentially; `loop.<name>`
  reads the previous iteration's container output `<name>` (always
  `.orValue(<seed>)` for the first pass). `max` resolves in parent scope,
  `until` in the loop container's own scope
- there is no `fail-fast`; keep loop/matrix state minimal and typed, and add
  explicit convergence, threshold, or stopping logic
- a root-level `finally:` is invalid; wrap the pipeline in a single container
  task that carries the `finally:`

## Validation

Templates are validated as a combined library with Wagie core, using Wagie's
executable preflight path (parse + compile). Two equivalent entry points:

```bash
make test                          # go test ./... — the full combined-library check
make validate                      # CLI report, every family
go run ./cmd/validate code         # report only files whose path contains "code"
go run ./cmd/validate ethereum/kurtosis-devnet-watch.yaml
```

`cmd/validate` always loads the whole library (so `uses:` composition resolves)
and prints one `ok`/`FAIL` line per family file; a path-substring filter only
narrows what is reported. Both paths validate against the pinned
`github.com/savid/wagie` release in `go.mod`.

The parser is strict: unknown fields are hard errors and it fails fast (one
error per file at a time), so re-run after each fix until `0 failed`.
`make validate` also prints non-failing topology warnings for risky optional
tasks, such as a task with `if:` that another sibling lists in `needs`. That
pattern is valid for deliberate branch pruning, but optional mainline stages
should usually run and return empty/default outputs instead.

Check templates in this order:

1. Compare against neighboring templates in the same family.
2. Check that the template still belongs in this repo and has not drifted into a generic core primitive.
3. Run `make validate` (or a filtered `go run ./cmd/validate <family>`) until green.
4. Run `make test` before finishing.

## Keep The Public Contract Small

A template should expose the smallest useful contract.

Prefer:

- clear input names
- stable outputs
- obvious control flow
- explicit descriptions on unusual requirements

Avoid hidden reliance on:

- undocumented side effects
- implicit worker behavior
- brittle output shapes that only make sense to one caller
