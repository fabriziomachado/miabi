// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package apply

import (
	"errors"
	"testing"

	d "github.com/miabi-io/miabi/internal/declarative"
	"github.com/miabi-io/miabi/internal/services/stack"
	"github.com/miabi-io/miabi/internal/storage/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func stackService(t *testing.T) *Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// Created by hand: the uid column's Postgres-only gen_random_uuid() default is not valid SQLite.
	if err := db.Exec(`CREATE TABLE stacks (id INTEGER PRIMARY KEY, uid TEXT, workspace_id INTEGER, name TEXT,
		display_name TEXT, docker_name TEXT, cluster_id INTEGER, docker_network TEXT, description TEXT,
		metadata TEXT, annotations TEXT, created_at DATETIME, updated_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	db.Exec(`INSERT INTO stacks (id, workspace_id, name, docker_name) VALUES (7, 1, 'shop', 'mb-shop'), (8, 2, 'shop', 'mb-shop-2')`)
	return &Service{stacks: stack.NewService(repositories.NewStackRepository(db), nil, nil, nil, nil, nil, nil, nil)}
}

// #451: spec.stack only steered placement, so the app never joined the stack it named.
func TestResolveStack(t *testing.T) {
	s := stackService(t)
	id, err := s.resolveStack(1, "web", &d.ApplicationSpec{Stack: "shop"})
	if err != nil || id == nil || *id != 7 {
		t.Fatalf("resolveStack = %v, %v; want the workspace's stack 7", id, err)
	}
	if id, err := s.resolveStack(1, "web", &d.ApplicationSpec{}); id != nil || err != nil {
		t.Errorf("no stack named should resolve to nil, got %v, %v", id, err)
	}
	if _, err := s.resolveStack(1, "web", &d.ApplicationSpec{Stack: "billing"}); !errors.Is(err, ErrInvalidManifest) {
		t.Errorf("an unknown stack should be a manifest error, got %v", err)
	}
}

// A single-app export carries no Stack resource, and a manifest may only name a Stack it declares.
func TestExportDropsStack(t *testing.T) {
	a := &d.ApplicationSpec{Stack: "shop"}
	trimDefaults(a, "")
	if a.Stack != "" {
		t.Errorf("exported stack = %q, want it dropped", a.Stack)
	}
}
