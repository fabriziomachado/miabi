// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package apply

import (
	"testing"

	d "github.com/miabi-io/miabi/internal/declarative"
	"github.com/miabi-io/miabi/internal/models"
)

func consoleApp() *models.Application {
	return &models.Application{
		HealthcheckType: models.HealthcheckHTTP, HealthcheckHTTPPath: "/healthz", HealthcheckPort: 8080,
		HealthcheckIntervalSeconds: 10, HealthcheckTimeoutSeconds: 5, HealthcheckRetries: 3,
	}
}

// Applying writes only what the manifest states, mirroring the diff.
func TestApplyHealthcheckKeepsUnstatedFields(t *testing.T) {
	app := consoleApp()
	applyHealthcheck(app, &d.HealthcheckSpec{Type: "http", Path: "/ready"})
	if app.HealthcheckHTTPPath != "/ready" || app.HealthcheckPort != 8080 || app.HealthcheckIntervalSeconds != 10 {
		t.Fatalf("app = %+v", app)
	}

	applyHealthcheck(app, nil)
	if app.HealthcheckType != models.HealthcheckHTTP {
		t.Fatal("a manifest without a healthcheck must leave the app's check alone")
	}

	applyHealthcheck(app, &d.HealthcheckSpec{Type: "none"})
	if app.HealthcheckType != models.HealthcheckNone {
		t.Fatalf("type = %s, want none", app.HealthcheckType)
	}
}

func TestHealthcheckSpecOfStatesType(t *testing.T) {
	if hc := healthcheckSpecOf(&models.Application{}); hc.Type != "none" {
		t.Errorf("an app without a check reads as type %q, want none", hc.Type)
	}
	hc := healthcheckSpecOf(consoleApp())
	if hc.Type != "http" || hc.Path != "/healthz" || hc.Port != 8080 || hc.IntervalSeconds != 10 {
		t.Errorf("spec = %+v", hc)
	}
}

func TestExportTrimsHealthcheckDefaults(t *testing.T) {
	a := &d.ApplicationSpec{Healthcheck: healthcheckSpecOf(consoleApp())}
	trimDefaults(a, "")
	want := d.HealthcheckSpec{Type: "http", Path: "/healthz", Port: 8080, IntervalSeconds: 10}
	if a.Healthcheck == nil || *a.Healthcheck != want {
		t.Errorf("exported healthcheck = %+v, want %+v", a.Healthcheck, want)
	}

	none := &d.ApplicationSpec{Healthcheck: healthcheckSpecOf(&models.Application{})}
	trimDefaults(none, "")
	if none.Healthcheck != nil {
		t.Error("an app without a check should export without a healthcheck block")
	}
}

func TestHealthcheckInput(t *testing.T) {
	if healthcheckInput(nil) != nil {
		t.Error("no manifest healthcheck means no create input")
	}
	in := healthcheckInput(&d.HealthcheckSpec{Type: "command", Command: "pg_isready", Retries: 5})
	if in.Type != models.HealthcheckCommand || in.Command != "pg_isready" || in.Retries != 5 {
		t.Errorf("input = %+v", in)
	}
}
