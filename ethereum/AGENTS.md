# ethereum/ — devnet workflows

Scope: the templates in `ethereum/`. Repo-wide policy lives in the root `AGENTS.md`.

## Division of labor

Templates here are the orchestration layer only: typed handoffs, fan-out, loops, and
gates. All Ethereum domain knowledge — procedure, vocabulary, judgment thresholds —
lives in the ethpandaops Panda runbooks and reaches workers at runtime through the
`panda` CLI. A template instruction never teaches domain procedure; it names which
runbook owns it and declares the typed output to come back with.

If a template seems to need domain knowledge no runbook owns, the fix is a runbook
change in panda, not more template prose.

## Runbook contracts are the source of truth

Template schemas mirror runbook output shapes. When they drift, change the template.
The owned shapes:

| Shape | Owning runbook |
|---|---|
| issue record, evidence item | `runbooks://devnet_issue_contract` |
| fingerprint block | `runbooks://devnet_issue_fingerprint_dedupe` |
| feedback queue | `runbooks://devnet_issue_feedback_queue` |
| root-cause report, reproduction status | `runbooks://devnet_issue_root_cause` |
| network_target | `runbooks://debug_ethereum_network` |
| watch window, service map, setup_summary | `runbooks://devnet_watch` |
| snapshot catalog entry (`epoch`, `snapshot_id`, `captured_at`) | `runbooks://panda_compute_kurtosis_lifecycle` |
| experiment triage disposition + campaign fields | `runbooks://devnet_issue_experiment_triage` |

Keep domain vocabulary out of schema `enum:` fields — name the values in the field
description with the owning runbook, so vocabulary changes don't need template
releases. Exception: the service `role` enum (`cl|el|vc|builder|tooling|unknown`) is
enforced because fingerprint component signatures depend on it.

## Instruction prose restates rules — the third drift channel

Schema mirrors and retrieval phrases are checked; the unchecked channel is
instruction prose that restates a runbook rule in miniature (a publish gate, a
carve-out summary, a snapshot-timing paraphrase). When the panda runbooks
change, sweep template instructions for stale restatements too — this is where
actively-wrong drift lives, not just missing vocabulary.

Two sanctioned patterns at the boundary, and only these:

- **Topology facts.** Runbook rules may branch on facts only the template
  knows — is this queue the only dispatch route? does this run own the
  sandbox? The template states the fact ("inside this loop the queue IS the
  dispatch route"); the rule that consumes it stays runbook-owned.
- **Labeled overrides.** A template may extend a runbook rule only when the
  orchestration layer genuinely needs it (e.g. carrying the full issue record
  in task inputs as the loop's transport), and must label the extension
  explicitly as orchestration-layer, citing the rule it extends. Keep these
  rare — every unlabeled deviation is a bug.

## Runbook references are retrieval-sensitive

Instructions name runbooks by meaning ("the runbook that owns collating watch
issues"); workers resolve them with `panda search runbooks "<need>"`. Before shipping
a phrase, verify it ranks the intended runbook first at the default limit:

```bash
panda search runbooks "root-cause a devnet issue"   # → devnet_issue_root_cause
panda read runbooks://devnet_issue_contract          # read a result
```

`make check-runbooks` verifies every concrete quoted phrase against the manifest in
`docs/runbook-refs.tsv` — add new phrases there, keep each on one line in template
instructions, and re-run it when the panda runbooks change. Prose-only references
("the runbook that owns …") are still unchecked.

## Validation

```bash
make validate FILTER=ethereum
```
