// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package serviceaccount

import (
	"errors"
	"strings"
	"testing"

	"github.com/miabi-io/miabi/internal/models"
	"github.com/miabi-io/miabi/internal/services/auth"
	"github.com/miabi-io/miabi/internal/storage/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestService(t *testing.T) (*Service, *repositories.WorkspaceRepository, *auth.APIKeyService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.WorkspaceMember{}, &models.APIKey{}); err != nil {
		t.Fatal(err)
	}
	users := repositories.NewUserRepository(db)
	workspaces := repositories.NewWorkspaceRepository(db)
	keyRepo := repositories.NewAPIKeyRepository(db)
	keys := auth.NewAPIKeyService(keyRepo)
	for _, m := range []models.WorkspaceMember{
		{WorkspaceID: 1, UserID: 100, Role: models.WorkspaceRoleAdmin},
		{WorkspaceID: 2, UserID: 200, Role: models.WorkspaceRoleAdmin},
	} {
		if err := workspaces.AddMember(&m); err != nil {
			t.Fatal(err)
		}
	}
	return NewService(users, workspaces, keys, keyRepo), workspaces, keys
}

func TestServiceAccountLifecycle(t *testing.T) {
	s, workspaces, keys := newTestService(t)

	if _, err := s.Create(1, "CI", models.WorkspaceRoleOwner); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("owner role accepted: %v", err)
	}
	sa, err := s.Create(1, "Deploy Bot", models.WorkspaceRoleDeveloper)
	if err != nil {
		t.Fatal(err)
	}
	if !sa.IsService() || !strings.HasSuffix(sa.Email, "@"+models.ServiceAccountEmailDomain) || !strings.HasPrefix(sa.Username, "sa-deploy-bot-") {
		t.Fatalf("unexpected account: %+v", sa)
	}
	if m, err := workspaces.FindMember(1, sa.ID); err != nil || m.Role != models.WorkspaceRoleDeveloper {
		t.Fatalf("membership: %+v %v", m, err)
	}

	if _, err := s.Get(2, sa.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("another workspace can see the account: %v", err)
	}

	plain, key, err := s.CreateKey(1, sa.ID, KeyRequest{Name: "ci", Scopes: []string{"deploy"}})
	if err != nil || key.WorkspaceID == nil || *key.WorkspaceID != 1 {
		t.Fatalf("key: %+v %v", key, err)
	}
	if _, err := keys.Verify(plain); err != nil {
		t.Fatalf("fresh key does not verify: %v", err)
	}

	if _, err := s.Delete(1, sa.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := keys.Verify(plain); err == nil {
		t.Fatal("key still verifies after the account was deleted")
	}
	if _, err := workspaces.FindMember(1, sa.ID); err == nil {
		t.Fatal("still a member of its workspace")
	}
	if list, _ := s.List(1); len(list) != 0 {
		t.Fatalf("deleted account still listed: %+v", list)
	}
}
