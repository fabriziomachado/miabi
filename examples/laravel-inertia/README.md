# Laravel + React (Inertia) — GitOps + Pipeline

End-to-end example of a **Laravel** app with a **React + Inertia.js** frontend,
**PostgreSQL**, and a persistent volume for **image uploads**, wired through
Miabi's **pipeline-as-code** and **GitOps** model.

```
laravel-inertia/
├── README.md                 # this file
├── Dockerfile                # sample multi-stage image (Composer + Vite + PHP)
├── .miabi/
│   └── pipeline.yaml         # test → build → (optional) deploy
└── envs/
    ├── dev/stack.yaml        # Postgres + uploads volume + app + route
    └── prod/stack.yaml       # same, digest-pinned, Swarm service
```

## How the pieces fit

| Piece | Role |
|---|---|
| **Pipeline** (`.miabi/pipeline.yaml`) | On `git push`, clones the commit, runs PHP/JS checks, builds the image with `uses: build`, and pushes it to Miabi's **built-in registry** |
| **GitOps** (`envs/*/stack.yaml`) | Declares the desired runtime: managed Postgres, uploads volume, Application (image + digest), Domain/Route |
| **Swarm** (`deployment.runtime: service` in prod) | Replicated tasks pull the same registry image; the control plane already pushed it there for multi-node deploys |

Promotion is a Git commit: pipeline produces `$MIABI_IMAGE_DIGEST` → you write
that digest into `envs/prod/stack.yaml` → the prod GitSource reconciles. A
revert commit rolls back.

> **Two modes.** Bind the pipeline to the **application** and keep `uses: deploy`
> if you want continuous deploy of that app (typical for `dev`). For **prod
> GitOps**, prefer build-only (or deploy only to a staging app) and promote by
> digest commit — otherwise the pipeline and the reconciler can race.

## Prerequisites

1. One **workspace per environment** (or only one of `envs/dev` / `envs/prod`
   synced into a given workspace) — resource names (`laravel-app`, `laravel-db`)
   match across folders on purpose, like the shop GitOps example.
2. A registered **Domain** covering the Route hosts (declared in the stacks, or
   created once under Networking → Domains).
3. The built-in **container registry** enabled (default on a full Miabi install).
4. A **Git repository** for the Laravel app (the pipeline's checkout source).
5. For prod Swarm: a location running **Docker Swarm** (`deployment.runtime: service`).

## 1. Put the pipeline in the app repo

Copy `.miabi/pipeline.yaml` and the sample `Dockerfile` into your Laravel repo
(adjust paths if your frontend lives elsewhere). When you create a Git-backed
application, Miabi can **adopt** `.miabi/pipeline.yaml` as a repo-owned pipeline.

Or register it yourself (bind to the app id):

```bash
BASE=https://miabi.example.com
WS=acme
APP=42
TOKEN=mb_xxx

curl -X POST "$BASE/api/v1/workspaces/$WS/pipelines" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "$(jq -Rs --argjson app "$APP" \
        '{name: "laravel-app", application_id: $app, spec: ., enabled: true}' \
        < .miabi/pipeline.yaml)"
```

A successful `uses: build` lands the image in the workspace registry, e.g.:

```
<registry-host>/ws_<workspace-id>/<app>:<run-number>
```

Browse it under **Container Registry**. Multi-node / Swarm nodes pull from there.

## 2. Wire GitOps per environment

Commit `envs/` to a repo (this tree, or yours). Create one GitSource per env:

```bash
# Dev — path envs/dev
curl -X POST "$BASE/api/v1/workspaces/$WS/gitops" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
        "name": "laravel-dev",
        "git_repository_id": 1,
        "ref": "main",
        "path": "examples/laravel-inertia/envs/dev",
        "sync_policy": "auto",
        "prune": true,
        "self_heal": true
      }'

# Prod — path envs/prod
curl -X POST "$BASE/api/v1/workspaces/$WS/gitops" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
        "name": "laravel-prod",
        "git_repository_id": 1,
        "ref": "main",
        "path": "examples/laravel-inertia/envs/prod",
        "sync_policy": "auto",
        "prune": true,
        "self_heal": true
      }'
```

Point `path` at wherever you keep these manifests in *your* repo.

## 3. Promote an image to prod

After a green pipeline run, pin the digest in `envs/prod/stack.yaml`:

```yaml
spec:
  image: registry.example.com/ws_1/laravel-app   # your registry + app repo
  tag: "42"                                      # optional human tag (run number)
  digest: "sha256:…"                             # $MIABI_IMAGE_DIGEST
```

Commit and push. The prod GitSource converges the Swarm service to that exact
artifact.

## What the stack declares

- **Database** `laravel-db` — managed PostgreSQL; credentials interpolated into
  `DB_*` env vars (and listed under `secretEnv`).
- **Volume** `laravel-uploads` — mounted at `/var/www/html/storage/app/public`
  for `FILESYSTEM_DISK=public` image uploads (and the storage symlink target).
- **Secret** `laravel-app-key` — documents a generated secret; set `APP_KEY` in
  the Application env before apply (apply/GitOps has no `{{ .secrets.* }}`
  helper — that is marketplace-only).
- **Application** `laravel-app` — image from the registry, HTTP healthcheck,
  rolling deploys; prod uses `runtime: service` for Swarm.
- **Domain + Route** — HTTPS via ACME on `laravel.example.com` /
  `dev.laravel.example.com`.

## Uploads and replicas

A **node-local** volume is fine for a single replica. If prod runs
`replicas: > 1` across Swarm nodes, every task needs the **same** files — use a
shared `storageClass` (NFS / CIFS / operator-managed host path under `/mnt/*`)
registered by the admin, or an object store (`FILESYSTEM_DISK=s3`) instead of a
local volume. Otherwise each replica sees a different disk.

## Sample Dockerfile notes

The included `Dockerfile` is illustrative: Composer install, `npm ci` +
`npm run build` (Vite / Inertia), then a PHP-FPM + Nginx runtime image. Point
`uses: build` at it (default `Dockerfile` at the repo root). Adapt the base
images and the document root to match your real Laravel layout.
