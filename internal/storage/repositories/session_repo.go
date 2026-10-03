// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package repositories

import (
	"github.com/miabi-io/miabi/internal/models"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository { return &SessionRepository{db: db} }

func (r *SessionRepository) Create(s *models.Session) error {
	return r.db.Create(s).Error
}

func (r *SessionRepository) FindByJTI(jti string) (*models.Session, error) {
	var s models.Session
	if err := r.db.Where("jti = ?", jti).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepository) RevokeByJTI(jti string) error {
	return r.db.Model(&models.Session{}).Where("jti = ?", jti).Update("revoked", true).Error
}

// RevokeAllForUser marks every session of a user revoked, except the one with exceptJTI (empty for all).
func (r *SessionRepository) RevokeAllForUser(userID uint, exceptJTI string) error {
	q := r.db.Model(&models.Session{}).Where("user_id = ? AND revoked = ?", userID, false)
	if exceptJTI != "" {
		q = q.Where("jti <> ?", exceptJTI)
	}
	return q.Update("revoked", true).Error
}

func (r *SessionRepository) ListByUser(userID uint) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

// FindByIDForUser loads a session by id scoped to its owner, so a user can only
// ever address their own sessions.
func (r *SessionRepository) FindByIDForUser(id, userID uint) (*models.Session, error) {
	var s models.Session
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}
