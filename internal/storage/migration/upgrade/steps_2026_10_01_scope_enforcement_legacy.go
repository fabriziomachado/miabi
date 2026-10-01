// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package upgrade

import (
	"context"
	"fmt"

	"github.com/miabi-io/miabi/internal/middlewares"
	"github.com/miabi-io/miabi/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// scopeEnforcementLegacyStep keeps an install that already issued API keys in warn mode: scopes
// were never enforced before, so those keys may rely on access they were never granted. New
// installs enforce from the start.
func scopeEnforcementLegacyStep(ctx context.Context, db *gorm.DB) error {
	var keys int64
	if err := db.WithContext(ctx).Table("api_keys").Where("revoked = ? AND ephemeral = ?", false, false).Count(&keys).Error; err != nil {
		return fmt.Errorf("count api keys: %w", err)
	}
	if keys == 0 {
		return nil
	}
	return db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.Setting{Key: middlewares.LegacyScopeModeSetting, Value: "warn", Type: models.SettingTypeString}).Error
}

func init() {
	steps = append(steps, Step{
		Name:    "api_key_scope_enforcement_legacy",
		Version: "1.10.11",
		Run:     scopeEnforcementLegacyStep,
	})
}
