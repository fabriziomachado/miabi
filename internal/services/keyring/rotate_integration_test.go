// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package keyring

import (
	"context"
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/miabi-io/miabi/internal/models"
	"github.com/miabi-io/miabi/internal/services/crypto"
	"github.com/miabi-io/miabi/internal/storage/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type strayRow struct {
	ID     uint `gorm:"primaryKey"`
	Secret string
	Doc    string `gorm:"type:jsonb"`
}

func (strayRow) TableName() string { return "h9_stray" }

type strayReencryptor struct{ db *gorm.DB }

func (r strayReencryptor) Reencrypt(_ context.Context, ws uint) (int, error) {
	var rows []strayRow
	if err := r.db.Find(&rows).Error; err != nil {
		return 0, err
	}
	for i := range rows {
		v, _, err := crypto.Reencrypt(ws, rows[i].Secret)
		if err != nil {
			return 0, err
		}
		pt, err := crypto.Decrypt(rows[i].Doc[len(`{"k": "`) : len(rows[i].Doc)-2])
		if err != nil {
			return 0, err
		}
		doc, err := crypto.EncryptWS(ws, pt)
		if err != nil {
			return 0, err
		}
		rows[i].Secret, rows[i].Doc = v, `{"k": "`+doc+`"}`
		if err := r.db.Save(&rows[i]).Error; err != nil {
			return 0, err
		}
	}
	return len(rows), nil
}

// A rotation must not delete key versions that unregistered owners' ciphertext still uses (H9).
func TestRotateKeepsOldVersionsWhileCiphertextReferencesThem(t *testing.T) {
	dsn := os.Getenv("MIABI_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set MIABI_TEST_POSTGRES_DSN to run the key rotation integration test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Migrator().DropTable(&strayRow{}, &models.WorkspaceKey{})
	if err := db.AutoMigrate(&models.WorkspaceKey{}, &strayRow{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Migrator().DropTable(&strayRow{}, &models.WorkspaceKey{}) })

	crypto.Init("test-master-key-for-keyring")
	s := NewService(repositories.NewWorkspaceKeyRepository(db))
	crypto.SetKeyring(s)
	t.Cleanup(func() { crypto.SetKeyring(nil) })

	const ws = 3
	secret, err := crypto.EncryptWS(ws, "s3-secret")
	if err != nil {
		t.Fatal(err)
	}
	inDoc, err := crypto.EncryptWS(ws, "passphrase")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&strayRow{Secret: secret, Doc: `{"k": "` + inDoc + `"}`}).Error; err != nil {
		t.Fatal(err)
	}

	res, err := s.Rotate(context.Background(), ws)
	if !errors.Is(err, ErrStaleCiphertext) {
		t.Fatalf("rotate with an unregistered owner: err = %v, want ErrStaleCiphertext", err)
	}
	if !slices.Equal(res.StaleColumns, []string{"h9_stray.doc", "h9_stray.secret"}) {
		t.Fatalf("stale columns = %v", res.StaleColumns)
	}
	if got, err := crypto.Decrypt(secret); err != nil || got != "s3-secret" {
		t.Fatalf("v1 ciphertext after rotation: %q, %v", got, err)
	}

	s.Register(strayReencryptor{db: db})
	res, err = s.Rotate(context.Background(), ws)
	if err != nil || len(res.StaleColumns) != 0 {
		t.Fatalf("rotate with every owner registered: res=%+v err=%v", res, err)
	}
	keys, err := s.repo.ListByWorkspace(ws)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0].Version != res.Version {
		t.Fatalf("keys after a clean rotation = %+v, want only v%d", keys, res.Version)
	}
}
