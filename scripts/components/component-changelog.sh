#!/usr/bin/env bash
# Lists the feat/fix/perf commits that touched a component's code, as a changelog draft.
#   scripts/components/component-changelog.sh gitops [since-ref]
# since-ref defaults to the last commit that changed the component's version constant.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
id="${1:?usage: $0 <component> [since-ref]}"
line=$(awk -v id="$id" '$1 == id' scripts/components/components.map)
[[ -z "$line" ]] && { echo "unknown component: $id" >&2; exit 1; }
read -r _ const paths <<<"$line"
since="${2:-$(git log -1 --format=%H -S"$const" -- internal/components/versions.go)}"
range="${since:+$since..}HEAD"
# shellcheck disable=SC2086
git log --no-merges --format='%h %s' "$range" -- $paths | grep -E '^[0-9a-f]+ (feat|fix|perf)(\(|!|:)' || echo "no feat/fix/perf changes since ${since:-the beginning}"
