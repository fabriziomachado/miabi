#!/usr/bin/env bash
# Installs the edge gateway on the manager node over the API, as the Admin Dashboard's action does.
source "$(dirname "$0")/lib.sh"
load

log "Log in and deploy the manager's gateway"
login "$ADMIN_EMAIL" "$ADMIN_PW"
MANAGER_ID=$(api GET /admin/nodes | jq -r '.data[] | select(.is_local) | .id')
[[ -n $MANAGER_ID ]] || fail "no local (manager) node listed"
save MANAGER_ID "$MANAGER_ID"
api POST "/admin/nodes/$MANAGER_ID/gateway/deploy" '{}' >/dev/null
wait_until "mb-node-gateway healthy" 180 healthy mb-node-gateway
