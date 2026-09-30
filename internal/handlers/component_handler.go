// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package handlers

import (
	"github.com/jkaninda/okapi"
	"github.com/miabi-io/miabi/internal/components"
)

// ComponentHandler lists the parts of Miabi that are versioned on their own.
type ComponentHandler struct {
	svc *components.Service
}

func NewComponentHandler(svc *components.Service) *ComponentHandler {
	return &ComponentHandler{svc: svc}
}

// List returns every component with its version, stability and status on this instance.
func (h *ComponentHandler) List(c *okapi.Context) error {
	return ok(c, h.svc.List())
}
