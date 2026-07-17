#!/usr/bin/env bash
# check-runbooks — verify that every retrieval phrase templates hand to
# `panda search runbooks "<phrase>"` still ranks its intended runbook first.
#
# Phrase -> expected-title pairs live in docs/runbook-refs.tsv. The script also
# scans the template families for quoted phrases missing from the manifest, so
# a new template cannot ship an unchecked reference. Requires the panda CLI.
set -euo pipefail

cd "$(dirname "$0")/.."
manifest="docs/runbook-refs.tsv"
families=(ethereum code research experiments)
fail=0

if ! command -v panda >/dev/null 2>&1; then
  echo "SKIP  panda CLI not available; cannot check runbook references" >&2
  exit 0
fi

# 1. Every manifest phrase must rank its expected runbook first.
while IFS=$'\t' read -r phrase expected; do
  [[ -z "$phrase" || "$phrase" == \#* ]] && continue
  # `|| true` guards the SIGPIPE (exit 141) panda takes under pipefail when it
  # keeps writing after head exits; without it set -e kills the whole check.
  first=$(panda search runbooks "$phrase" 2>/dev/null | head -1 || true)
  if [[ "$first" == "$expected"* ]]; then
    echo "ok    \"$phrase\" -> $expected"
  else
    echo "FAIL  \"$phrase\" -> got: ${first:-<no result>}; want: $expected"
    fail=1
  fi
done < "$manifest"

# 2. Every quoted phrase used in a template must be in the manifest.
# Whitespace is normalized and YAML backslash-escaped quotes are unescaped
# first, so neither folded lines nor phrases embedded in double-quoted
# scalars (e.g. inside CEL expressions) can hide a phrase from the scan.
used=$(for f in $(find "${families[@]}" -name '*.yaml' 2>/dev/null); do
  tr '\n' ' ' < "$f" | sed 's/  */ /g; s/\\"/"/g'
done | { grep -oP 'panda search runbooks "\K[^"]+' || true; } | { grep -v '<' || true; } | sort -u)

while IFS= read -r phrase; do
  [[ -z "$phrase" ]] && continue
  if ! awk -F'\t' -v p="$phrase" '$1 == p { found = 1 } END { exit !found }' "$manifest"; then
    echo "MISS  phrase used in templates but not in $manifest: \"$phrase\""
    fail=1
  fi
done <<< "$used"

# 3. Every concrete direct-read runbook ref in a template must resolve.
# (Placeholders like runbooks://<stem> are excluded by the character class.)
refs=$(for f in $(find "${families[@]}" -name '*.yaml' 2>/dev/null); do
  cat "$f"
done | { grep -oP 'panda read runbooks://\K[a-z0-9_]+' || true; } | sort -u)

while IFS= read -r stem; do
  [[ -z "$stem" ]] && continue
  if panda read "runbooks://$stem" >/dev/null 2>&1; then
    echo "ok    runbooks://$stem resolves"
  else
    echo "FAIL  runbooks://$stem is referenced in templates but not readable (undeployed or renamed runbook)"
    fail=1
  fi
done <<< "$refs"

exit "$fail"
