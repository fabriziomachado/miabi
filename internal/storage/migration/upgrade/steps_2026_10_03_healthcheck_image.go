// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package upgrade

import (
	"context"

	"gorm.io/gorm"
)

// healthcheckImageStep moves apps on "none" to "image". "none" used to keep the image's own
// HEALTHCHECK and now disables it, so "image" is what those apps were running with.
func healthcheckImageStep(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Table("applications").
		Where("healthcheck_type = ? OR healthcheck_type = ''", "none").
		Update("healthcheck_type", "image").Error
}

func init() {
	steps = append(steps, Step{
		Name:    "healthcheck_none_to_image",
		Version: "1.10.12",
		Run:     healthcheckImageStep,
	})
}
