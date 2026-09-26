#!/usr/bin/env bash
# Prints what a failed run needs to be diagnosed without a rerun: every container's state and the logs of
# the platform's own.
docker ps -a --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
for c in miabi mb-node-gateway miabi-gateway miabi-postgres miabi-redis; do
  if docker inspect "$c" >/dev/null 2>&1; then
    printf '\n===== %s =====\n' "$c"
    docker logs --tail 300 "$c" 2>&1
  fi
done
