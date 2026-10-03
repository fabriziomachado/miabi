// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package declarative_test

import (
	"strings"
	"testing"

	d "github.com/miabi-io/miabi/internal/declarative"
	"github.com/miabi-io/miabi/internal/models"
)

func hcAppYAML(healthcheck string) string {
	body := `
apiVersion: miabi.io/v1
kind: Application
metadata: { name: web }
spec:
  image: ghcr.io/org/web
  ports: [{ container: 8080 }]
`
	if healthcheck != "" {
		body += "  healthcheck:\n" + healthcheck
	}
	return body
}

func TestHealthcheckParses(t *testing.T) {
	set, err := d.Parse([]byte(hcAppYAML(`
    type: http
    path: /healthz
    port: 8080
    intervalSeconds: 10
    startPeriodSeconds: 20
`)))
	if err != nil {
		t.Fatal(err)
	}
	r, _ := set.Get("Application/web")
	hc := r.Application.Healthcheck
	if hc == nil || hc.Type != "http" || hc.Path != "/healthz" || hc.Port != 8080 || hc.IntervalSeconds != 10 || hc.StartPeriodSeconds != 20 {
		t.Fatalf("healthcheck = %+v", hc)
	}
}

func TestHealthcheckValidation(t *testing.T) {
	cases := map[string]string{
		"    type: tcp\n":                     "must be http, command or none",
		"    path: /healthz\n":                "must be http, command or none",
		"    type: command\n":                 "healthcheck.command is required",
		"    type: http\n    path: healthz\n": "must start with /",
		"    type: http\n    port: 70000\n":   "between 1 and 65535",
		"    type: http\n    retries: -1\n":   "cannot be negative",
	}
	for hc, want := range cases {
		if _, err := d.Parse([]byte(hcAppYAML(hc))); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: error = %v, want it to mention %q", hc, err, want)
		}
	}
}

// The declarative package keeps its own list of types; this stops it drifting from the model's.
func TestManifestHealthcheckTypesMatchTheModel(t *testing.T) {
	for _, typ := range []models.HealthcheckType{models.HealthcheckNone, models.HealthcheckHTTP, models.HealthcheckCommand} {
		body := "    type: " + string(typ) + "\n"
		if typ == models.HealthcheckCommand {
			body += "    command: pg_isready\n"
		}
		if _, err := d.Parse([]byte(hcAppYAML(body))); err != nil {
			t.Errorf("the model accepts %q but the manifest does not: %v", typ, err)
		}
	}
}

// liveHC is the live side as the apply engine renders it: every field stated.
const liveHC = `
    type: http
    path: /healthz
    port: 8080
    intervalSeconds: 30
    timeoutSeconds: 5
    retries: 3
`

func healthcheckDiffs(t *testing.T, desiredHC, actualHC string) map[string]d.FieldDiff {
	t.Helper()
	desired, err := d.Parse([]byte(hcAppYAML(desiredHC)))
	if err != nil {
		t.Fatal(err)
	}
	actual, err := d.Parse([]byte(hcAppYAML(actualHC)))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]d.FieldDiff{}
	for _, c := range d.BuildPlan(desired, actual, d.PlanOptions{}).Changes {
		for _, f := range c.Fields {
			out[f.Field] = f
		}
	}
	return out
}

// A manifest silent about the healthcheck must not reset one configured in the console.
func TestOmittedHealthcheckIsNotDrift(t *testing.T) {
	if diffs := healthcheckDiffs(t, "", liveHC); len(diffs) != 0 {
		t.Errorf("a manifest without a healthcheck planned %v", diffs)
	}
}

// Only the stated fields are compared: stating type and path leaves the console's timing alone.
func TestPartialHealthcheckComparesStatedFields(t *testing.T) {
	if diffs := healthcheckDiffs(t, "    type: http\n    path: /healthz\n", liveHC); len(diffs) != 0 {
		t.Errorf("a manifest matching on its stated fields planned %v", diffs)
	}
}

func TestDeclaredHealthcheckConverges(t *testing.T) {
	diffs := healthcheckDiffs(t, "    type: http\n    path: /ready\n    retries: 5\n", liveHC)
	if f := diffs["healthcheck.path"]; f.From != "/healthz" || f.To != "/ready" {
		t.Errorf("path diff = %+v", f)
	}
	if f := diffs["healthcheck.retries"]; f.From != "3" || f.To != "5" {
		t.Errorf("retries diff = %+v", f)
	}
	if _, ok := diffs["healthcheck.intervalSeconds"]; ok {
		t.Error("an unstated interval must not be compared")
	}
}

// type: none is how a manifest turns a check off.
func TestHealthcheckNoneDisables(t *testing.T) {
	if f := healthcheckDiffs(t, "    type: none\n", liveHC)["healthcheck.type"]; f.From != "http" || f.To != "none" {
		t.Errorf("type diff = %+v", f)
	}
}
