// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package routes

import (
	"net/http"

	"github.com/jkaninda/okapi"
	"github.com/miabi-io/miabi/internal/components"
	"github.com/miabi-io/miabi/internal/dto"
)

// componentRoutes serve the independently versioned components shown on the About page.
func (r *Router) componentRoutes() []okapi.RouteDefinition {
	sys := r.v1.Group("/system").WithTagInfo(okapi.GroupTag{Name: "Components", Description: "Independently versioned parts of Miabi."})
	return []okapi.RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "/components",
			Group:       sys,
			Middlewares: []okapi.Middleware{r.authenticate},
			Handler:     r.h.components.List,
			Summary:     "Component versions, stability and status (any signed-in user)",
			Response:    &dto.Response[[]components.Component]{},
		},
	}
}
