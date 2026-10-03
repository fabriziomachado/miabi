// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package declarative_test

import (
	"testing"

	d "github.com/miabi-io/miabi/internal/declarative"
)

func stackBundle(appStack string) string {
	body := `
apiVersion: miabi.io/v1
kind: Stack
metadata: { name: shop }
---
apiVersion: miabi.io/v1
kind: Application
metadata: { name: web }
spec:
  image: ghcr.io/org/web
`
	if appStack != "" {
		body += "  stack: " + appStack + "\n"
	}
	return body
}

func stackDiff(t *testing.T, desired, actual string) (d.FieldDiff, bool) {
	t.Helper()
	ds, err := d.Parse([]byte(desired))
	if err != nil {
		t.Fatal(err)
	}
	as, err := d.Parse([]byte(actual))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range d.BuildPlan(ds, as, d.PlanOptions{}).Changes {
		for _, f := range c.Fields {
			if f.Field == "stack" {
				return f, true
			}
		}
	}
	return d.FieldDiff{}, false
}

// Joining a stack from a manifest has to plan as a change of the app.
func TestStatedStackConverges(t *testing.T) {
	f, ok := stackDiff(t, stackBundle("shop"), stackBundle(""))
	if !ok || f.From != "" || f.To != "shop" {
		t.Errorf("stack diff = %+v (planned %v), want \"\" -> shop", f, ok)
	}
}

// A manifest silent about the stack leaves a membership set in the console alone.
func TestOmittedStackIsNotDrift(t *testing.T) {
	if f, ok := stackDiff(t, stackBundle(""), stackBundle("shop")); ok {
		t.Errorf("a manifest without a stack planned %+v", f)
	}
}
