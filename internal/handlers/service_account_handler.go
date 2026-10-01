// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/jkaninda/okapi"
	"github.com/miabi-io/miabi/internal/middlewares"
	"github.com/miabi-io/miabi/internal/models"
	"github.com/miabi-io/miabi/internal/services/audit"
	"github.com/miabi-io/miabi/internal/services/auth"
	"github.com/miabi-io/miabi/internal/services/serviceaccount"
)

// ServiceAccountHandler serves a workspace's service accounts and their API keys.
type ServiceAccountHandler struct {
	svc   *serviceaccount.Service
	audit *audit.Logger
}

func NewServiceAccountHandler(svc *serviceaccount.Service, auditLog *audit.Logger) *ServiceAccountHandler {
	return &ServiceAccountHandler{svc: svc, audit: auditLog}
}

type CreateServiceAccountRequest struct {
	Body struct {
		Name string               `json:"name" required:"true" maxLength:"64"`
		Role models.WorkspaceRole `json:"role" enum:"viewer,developer,admin" default:"developer"`
	} `json:"body"`
}

type CreateServiceAccountKeyRequest struct {
	Body struct {
		Name          string   `json:"name" required:"true"`
		Scopes        []string `json:"scopes" enum:"read,write,deploy,admin,*,registry_read,registry_write" default:"read"`
		AllowedIPs    []string `json:"allowed_ips"`
		ExpiresInDays *int     `json:"expires_in_days" min:"0"`
	} `json:"body"`
}

func (h *ServiceAccountHandler) record(c *okapi.Context, action, targetType string, targetID uint, meta map[string]any) {
	actor, ws := middlewares.UserID(c), middlewares.WorkspaceID(c)
	h.audit.Record(audit.Entry{ActorID: &actor, WorkspaceID: &ws, Action: action, TargetType: targetType, TargetID: strconv.Itoa(int(targetID)), IP: c.RealIP(), Metadata: meta})
}

func (h *ServiceAccountHandler) abort(c *okapi.Context, err error) error {
	switch {
	case errors.Is(err, serviceaccount.ErrNotFound), errors.Is(err, serviceaccount.ErrKeyNotFound):
		return c.AbortNotFound(err.Error())
	case errors.Is(err, serviceaccount.ErrInvalidName), errors.Is(err, serviceaccount.ErrInvalidRole):
		return c.AbortBadRequest(err.Error())
	}
	if a := quotaAbort(c, err); a != nil {
		return a
	}
	return c.AbortInternalServerError("service account operation failed", err)
}

func pathID(c *okapi.Context, name string) (uint, error) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		return 0, c.AbortBadRequest("invalid " + name)
	}
	return uint(id), nil
}

// Create adds a service account to the workspace.
func (h *ServiceAccountHandler) Create(c *okapi.Context, req *CreateServiceAccountRequest) error {
	u, err := h.svc.Create(middlewares.WorkspaceID(c), req.Body.Name, req.Body.Role)
	if err != nil {
		return h.abort(c, err)
	}
	h.record(c, "service_account.create", "user", u.ID, map[string]any{"name": u.Name, "role": req.Body.Role})
	return created(c, u)
}

// List returns the service accounts the workspace owns.
func (h *ServiceAccountHandler) List(c *okapi.Context) error {
	users, err := h.svc.List(middlewares.WorkspaceID(c))
	if err != nil {
		return h.abort(c, err)
	}
	return ok(c, users)
}

// Delete disables a service account, revoking its keys and memberships.
func (h *ServiceAccountHandler) Delete(c *okapi.Context) error {
	id, err := pathID(c, "serviceAccountID")
	if err != nil {
		return err
	}
	u, err := h.svc.Delete(middlewares.WorkspaceID(c), id)
	if err != nil {
		return h.abort(c, err)
	}
	h.record(c, "service_account.delete", "user", u.ID, map[string]any{"name": u.Name})
	return message(c, "service account deleted")
}

// CreateKey mints an API key for a service account.
func (h *ServiceAccountHandler) CreateKey(c *okapi.Context, req *CreateServiceAccountKeyRequest) error {
	id, err := pathID(c, "serviceAccountID")
	if err != nil {
		return err
	}
	scopes, err := auth.NormalizeScopes(req.Body.Scopes)
	if err != nil {
		return c.AbortBadRequest("invalid scope", err)
	}
	var expiresAt *time.Time
	if req.Body.ExpiresInDays != nil && *req.Body.ExpiresInDays > 0 {
		t := time.Now().AddDate(0, 0, *req.Body.ExpiresInDays)
		expiresAt = &t
	}
	plaintext, key, err := h.svc.CreateKey(middlewares.WorkspaceID(c), id, serviceaccount.KeyRequest{
		Name: req.Body.Name, Scopes: scopes, AllowedIPs: req.Body.AllowedIPs, ExpiresAt: expiresAt,
	})
	if err != nil {
		return h.abort(c, err)
	}
	h.record(c, "service_account.key_create", "api_key", key.ID, map[string]any{"service_account_id": id, "scopes": key.Scopes})
	return created(c, APIKeyCreated{
		ID: key.ID, Name: key.Name, Key: plaintext, KeyPrefix: key.KeyPrefix, Scopes: key.Scopes,
		AllowedIPs: key.AllowedIPs, WorkspaceID: key.WorkspaceID, ExpiresAt: key.ExpiresAt,
		Message: "Save this key securely. It will not be shown again.",
	})
}

// ListKeys returns a service account's API keys, without secrets.
func (h *ServiceAccountHandler) ListKeys(c *okapi.Context) error {
	id, err := pathID(c, "serviceAccountID")
	if err != nil {
		return err
	}
	keys, err := h.svc.ListKeys(middlewares.WorkspaceID(c), id)
	if err != nil {
		return h.abort(c, err)
	}
	return ok(c, keys)
}

// RevokeKey revokes one of a service account's API keys.
func (h *ServiceAccountHandler) RevokeKey(c *okapi.Context) error {
	id, err := pathID(c, "serviceAccountID")
	if err != nil {
		return err
	}
	keyID, err := pathID(c, "keyID")
	if err != nil {
		return err
	}
	k, err := h.svc.RevokeKey(middlewares.WorkspaceID(c), id, keyID)
	if err != nil {
		return h.abort(c, err)
	}
	h.record(c, "service_account.key_revoke", "api_key", k.ID, map[string]any{"service_account_id": id})
	return message(c, "API key revoked")
}
