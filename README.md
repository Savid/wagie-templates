# wagie-templates

Specialized Wagie template families live here.

Current layout:

- `ethereum`: devnet and Kurtosis Ethereum workflows
- `code`: code-review and code-quality workflows
- `ci`: CI failure triage and GitHub issue state workflows
- `research`: auto-research workflows
- `experiments`: metric-driven experiment loops over git-tracked artifacts

Validation is run as a combined library with Wagie core templates, through
Wagie's executable preflight (parse + compile), against the pinned
`github.com/savid/wagie` release in `go.mod`.

Useful commands:

```bash
make test                      # full combined-library validation
make validate                  # per-file ok/FAIL report across all families
make validate FILTER=ethereum  # report only paths containing "ethereum"
make tidy
```

See `docs/agent-guides/project-templates.md` for the spec grammar and authoring
guidance.
