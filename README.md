# wagie-templates

Specialized Wagie template families live here.

Families are shallow root-level directories, one per domain (`ethereum`, `code`,
`research`, `experiments`, …) — see each family's templates for what it covers.

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
