// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package declarative

import "testing"

// The check's field lists must match what the renderer resolves, or a dry run would reject a reference
// apply accepts (or wave through one it rejects).
func TestRefFieldsMatchRenderer(t *testing.T) {
	r := NewRenderer(RenderContext{
		Databases: map[string]ConnView{"db": {Host: "h"}},
		Apps:      map[string]AppView{"web": {Host: "web"}},
	})
	for _, f := range databaseRefFields {
		if _, err := r.ref("databases", "db", f); err != nil {
			t.Errorf("database field %q: the check allows it but the renderer refuses: %v", f, err)
		}
	}
	for _, f := range applicationRefFields {
		if _, err := r.ref("applications", "web", f); err != nil {
			t.Errorf("application field %q: the check allows it but the renderer refuses: %v", f, err)
		}
	}
	if _, err := r.ref("databases", "db", "uri2"); err == nil {
		t.Error("the renderer must refuse a field the check would reject")
	}
}
