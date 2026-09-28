// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package middleware

import (
	"context"

	"github.com/miabi-io/miabi/internal/mwcatalog"
)

// Reencrypt re-encrypts the secrets in the workspace's middleware rules under the workspace's active
// DEK (key rotation). Idempotent; returns the number of middlewares rewritten. The gateway needs no
// resync: the decrypted values it renders are unchanged.
func (s *Service) Reencrypt(ctx context.Context, workspaceID uint) (int, error) {
	rows, err := s.repo.ListByWorkspace(workspaceID)
	if err != nil {
		return 0, err
	}
	n := 0
	for i := range rows {
		rule, changed, rerr := mwcatalog.ReencryptSecrets(rows[i].Type, workspaceID, rows[i].Rule)
		if rerr != nil {
			return n, rerr
		}
		if changed {
			rows[i].Rule = rule
			if uerr := s.repo.Update(&rows[i]); uerr != nil {
				return n, uerr
			}
			n++
		}
	}
	return n, nil
}
