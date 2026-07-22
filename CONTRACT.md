# Adopting the wagie task `contract` field

Tracking doc for the changes this repo makes when wagie's `plan/contract/`
lands (a leaf-task `contract:` block declaring `mutable`/`frozen` workspace
path boundaries, enforced by the harness — see `wagie/plan/contract/README.md`).
Changes are staged: phase 0 is pure YAML and needs no spec support; phase 1
activates the field.

## Phase 0 — typed `frozen_files` outputs (no spec dependency — landed)

The scoring surface is currently implicit: prose in instructions plus
experiment-verify tracing what the scoring commands load. Make it a typed
output so it can be checked against — and so phase 1 has data to bind.

- `experiments/experiment-loop.yaml` — `setup` gains a required
  `frozen_files` output: every path the scoring contract reads (benchmark
  files, fixtures/datasets, eval splits, correctness test files), whether
  discovered or scaffolded. Exposed as a template output alongside the other
  `effective_*` values.
- `experiments/experiment-islands.yaml` — the `contract` task gains the same
  `frozen_files` output (the shared list every cell runs under); threaded to
  the `verify` binding.
- `experiments/experiment-verify.yaml` — new optional input `frozen_files`
  (default `[]`). `legality` consumes it as the declared scoring surface:
  classification checks the declared list first, and command-tracing becomes
  the backstop that catches undeclared reads (a scoring file missing from
  `frozen_files` is itself a finding worth reporting in `violations` detail).

## Phase 1 — bind contracts (needs the spec field + optimize-loop inputs)

Once core `optimize-loop` grows `mutable_paths`/`frozen_paths` inputs
(`wagie/plan/contract/03-templates.md`):

- `experiments/experiment-loop.yaml` — the `optimize` binding wires
  `effective_target_files` → `mutable_paths` and setup's `frozen_files` →
  `frozen_paths`. That puts the enforced contract on the propose/measure/enact
  leaves without any further change here.
- `experiments/experiment-islands.yaml` — nothing extra; cells inherit
  through experiment-loop.
- `experiments/experiment-verify.yaml` — `legality` declares a write-nothing
  contract (pure audit task). Blocked on the empty-array semantics decision
  (`wagie/plan/contract/README.md`): "write nothing" must be expressible
  without colliding with "empty list = boundary not declared".
- `experiments/experiment-loop.yaml` `finalize` — optional defense-in-depth:
  `frozen: frozen_files` (it writes reports and charts, never scoring files).

Not adopted, deliberately:

- `setup` (loop) and `contract` (islands) tasks: they *author* the scoring
  surface — a contract on the task that creates the frozen files is circular.
- `replicate`/`holdout` (verify): they rewrite the working tree by design
  (checkout per side, rebuilds), so a diff-based write boundary does not fit;
  their honest invariant is "no net new commits", which the contract field
  does not express. Revisit if enforcement grows that shape.
- Reasoning-only tasks (`curate`, `report`, `claim`, `decide`): no workspace,
  nothing to declare.

## Later candidates outside experiments

- `code/code-review-fix-loop.yaml` — same shape as the optimizer: mutable =
  the source being fixed, frozen = the tests backing the review verdict.
- `ethereum/` — mostly orchestrates network state rather than editing scored
  files; no candidate today.

## Retire this file

Fold the outcome into each template and delete this doc once phase 1 lands —
per repo rules, template contracts live in the templates, not in inventories.
