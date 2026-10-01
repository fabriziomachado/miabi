// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package routes

import (
	"net/http"

	"github.com/jkaninda/okapi"
	"github.com/miabi-io/miabi/internal/dto"
	"github.com/miabi-io/miabi/internal/handlers"
	"github.com/miabi-io/miabi/internal/middlewares"
	"github.com/miabi-io/miabi/internal/models"
)

// serviceAccountRoutes registers a workspace's service accounts and their keys. Admins only, and
// every change takes a signed-in session: like minting a personal key, a key must not be able to
// mint a longer-lived or wider one for a machine identity.
func (r *Router) serviceAccountRoutes() []okapi.RouteDefinition {
	g := r.v1.Group("/workspaces/{workspace}/service-accounts").WithTagInfo(okapi.GroupTag{Name: "Service accounts", Description: "Non-human workspace members that authenticate with API keys."})
	read := []okapi.Middleware{r.authenticate, r.scope, middlewares.RequireRole(models.WorkspaceRoleAdmin)}
	manage := []okapi.Middleware{r.authenticate, middlewares.RequireSession(), r.scope, middlewares.RequireRole(models.WorkspaceRoleAdmin)}

	return []okapi.RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "",
			Group:       g,
			Middlewares: read,
			Handler:     r.h.serviceAccount.List,
			Summary:     "List service accounts",
			Response:    &dto.Response[[]models.User]{},
		},
		{
			Method:      http.MethodPost,
			Path:        "",
			Group:       g,
			Middlewares: manage,
			Handler:     okapi.H(r.h.serviceAccount.Create),
			Summary:     "Create a service account",
			Request:     &handlers.CreateServiceAccountRequest{},
			Response:    &dto.Response[models.User]{},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/{serviceAccountID}",
			Group:       g,
			Middlewares: manage,
			Handler:     r.h.serviceAccount.Delete,
			Summary:     "Delete a service account",
			Response:    &dto.Response[dto.MessageData]{},
		},
		{
			Method:      http.MethodGet,
			Path:        "/{serviceAccountID}/keys",
			Group:       g,
			Middlewares: read,
			Handler:     r.h.serviceAccount.ListKeys,
			Summary:     "List a service account's API keys",
			Response:    &dto.Response[[]models.APIKey]{},
		},
		{
			Method:      http.MethodPost,
			Path:        "/{serviceAccountID}/keys",
			Group:       g,
			Middlewares: manage,
			Handler:     okapi.H(r.h.serviceAccount.CreateKey),
			Summary:     "Create an API key for a service account",
			Request:     &handlers.CreateServiceAccountKeyRequest{},
			Response:    &dto.Response[handlers.APIKeyCreated]{},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/{serviceAccountID}/keys/{keyID}",
			Group:       g,
			Middlewares: manage,
			Handler:     r.h.serviceAccount.RevokeKey,
			Summary:     "Revoke a service account's API key",
			Response:    &dto.Response[dto.MessageData]{},
		},
	}
}
