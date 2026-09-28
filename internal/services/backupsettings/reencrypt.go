// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package backupsettings

import (
	"context"
	"errors"

	"github.com/miabi-io/miabi/internal/services/crypto"
	"gorm.io/gorm"
)

// Reencrypt re-encrypts the workspace's S3 secret key and bundle/backup passphrases under the
// workspace's active DEK (key rotation). Idempotent; returns the number of settings rows rewritten.
func (s *Service) Reencrypt(ctx context.Context, workspaceID uint) (int, error) {
	st, err := s.repo.FindByWorkspace(workspaceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	changed := false
	for _, f := range []*string{&st.S3SecretKeyEnc, &st.BundlePassphraseEnc, &st.BackupPassphraseEnc} {
		v, ch, rerr := crypto.Reencrypt(workspaceID, *f)
		if rerr != nil {
			return 0, rerr
		}
		if ch {
			*f = v
			changed = true
		}
	}
	if !changed {
		return 0, nil
	}
	if err := s.repo.Upsert(st); err != nil {
		return 0, err
	}
	return 1, nil
}
