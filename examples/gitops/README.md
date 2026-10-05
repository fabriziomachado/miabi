# GitOps example

A monorepo layout with **per-environment folders** (recommended over
branch-per-env — diffs and promotion PRs are clearer):

```
envs/
├── dev/stack.yaml    # database + volume + app (tag: dev) + route
└── prod/stack.yaml   # same, digest-pinned, larger resources, prod host
```

Each folder is a complete desired state. A **GitSource** points at one folder
and Miabi continuously converges the workspace to it:
`Git → (sync) → desired state → (reconciler) → Docker`.

## Setup

The calls below assume these shell variables. Address the workspace by its
**name** (its handle) — the `/workspaces/{ws}` path accepts the name, its uid, or
the numeric id:

```bash
BASE=https://miabi.example.com   # your Miabi URL
WS=acme                          # workspace name (handle) — numeric id also works
TOKEN=mb_xxx                     # API token (Settings → API keys)
ID=1                             # a GitSource id (for sync/diff below)
```

## Wire it up

Create one GitSource per environment (path = the env folder):

```bash
curl -X POST "$BASE/api/v1/workspaces/$WS/gitops" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
        "name": "prod",
        "git_repository_id": 1,
        "ref": "main",
        "path": "envs/prod",
        "sync_policy": "auto",
        "prune": true,
        "self_heal": true
      }'
```

- **git_repository_id** — references a Git Repository (Networking-adjacent
  *Sources → Git Repositories*). The repo URL and clone credentials come from it;
  add a Git Repository first (it may be **public** — leave the token blank — or
  private). In the web UI you just pick it from a dropdown. (A raw `repo_url` is
  still accepted by the API as an alternative.)
- **sync_policy: auto** — reconciles on the 3-minute sweep and on push webhook.
- **prune** — deletes resources removed from Git (only ones GitOps created;
  hand-made resources are never pruned).
- **self_heal** — re-applies when live state drifts from Git.

Trigger a sync now, or inspect the desired-vs-live diff:

```bash
curl -X POST "$BASE/api/v1/workspaces/$WS/gitops/$ID/sync"  -H "Authorization: Bearer $TOKEN"
curl      "$BASE/api/v1/workspaces/$WS/gitops/$ID/diff"     -H "Authorization: Bearer $TOKEN"
```

## Push webhook

Add the source's generated secret as a GitHub `X-Hub-Signature-256` webhook (or
GitLab `X-Gitlab-Token`) pointing at:

```
POST $BASE/api/v1/workspaces/$WS/gitops/$ID/webhook
```

A verified push triggers an immediate sync.

## Promotion

Promote dev → prod by committing the tested image's digest into
`envs/prod/stack.yaml` (a revert commit rolls back). The prod GitSource
reconciles the change — auditable and reproducible.

For a full stack that pairs this promotion model with a **pipeline build** into
Miabi's registry (Laravel + React/Inertia, Postgres, upload volume, Swarm in
prod), see [../laravel-inertia/](../laravel-inertia/).
