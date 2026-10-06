// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package upgrade

import (
	"context"

	"github.com/miabi-io/miabi/internal/models"
	"github.com/miabi-io/miabi/internal/services/database"
	"gorm.io/gorm"
)

// dbURLEnvStep adds DB_URL next to the DATABASE_URL an attached database
// injected, now that DATABASE_URL is the deprecated alias. Only a DATABASE_URL
// that still points at the database's own URL secret is copied, so a value the
// user or a template wrote by hand is left alone.
func dbURLEnvStep(ctx context.Context, db *gorm.DB) error {
	tx := db.WithContext(ctx)
	var dbs []models.Database
	if err := tx.Where("application_id IS NOT NULL").Find(&dbs).Error; err != nil {
		return err
	}
	for i := range dbs {
		d := &dbs[i]
		var inst models.DatabaseInstance
		if err := tx.Select("id", "name", "engine").First(&inst, d.InstanceID).Error; err != nil {
			continue
		}
		legacy, url := "DATABASE_URL", "DB_URL"
		if d.EnvPrefix != "" {
			legacy, url = d.EnvPrefix+"_"+legacy, d.EnvPrefix+"_"+url
		}
		ref := "${{ secrets." + database.URLSecretName(&inst, d) + " }}"
		var src models.AppEnvVar
		if err := tx.Where("application_id = ? AND key = ? AND value = ?", *d.ApplicationID, legacy, ref).
			First(&src).Error; err != nil {
			continue
		}
		var n int64
		if err := tx.Model(&models.AppEnvVar{}).Where("application_id = ? AND key = ?", *d.ApplicationID, url).
			Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		if err := tx.Create(&models.AppEnvVar{ApplicationID: *d.ApplicationID, Key: url, Value: ref}).Error; err != nil {
			return err
		}
	}
	return nil
}

func init() {
	steps = append(steps, Step{
		Name:    "db_url_env_alias",
		Version: "1.10.13",
		Run:     dbURLEnvStep,
	})
}
