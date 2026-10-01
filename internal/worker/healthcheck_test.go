// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package worker

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/miabi-io/miabi/internal/docker"
	"github.com/miabi-io/miabi/internal/healthprobe"
	"github.com/miabi-io/miabi/internal/models"
)

func TestHTTPCheckTarget(t *testing.T) {
	cases := []struct {
		app      models.Application
		port     int
		wantPath string
	}{
		{models.Application{}, 80, "/"},
		{models.Application{Port: 3000, HealthcheckHTTPPath: "healthz"}, 3000, "/healthz"},
		{models.Application{Port: 3000, HealthcheckPort: 9090, HealthcheckHTTPPath: " /ready "}, 9090, "/ready"},
	}
	for _, c := range cases {
		port, path := httpCheckTarget(&c.app)
		if port != c.port || path != c.wantPath {
			t.Errorf("httpCheckTarget(%+v) = %d %q, want %d %q", c.app, port, path, c.port, c.wantPath)
		}
	}
}

func TestShellQuote(t *testing.T) {
	for _, in := range []string{"/healthz", "/x';touch /tmp/pwned;'", `/a b"$HOME"` + "`id`"} {
		out, err := exec.Command("sh", "-c", "printf %s "+shellQuote(in)).Output()
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != in {
			t.Errorf("shellQuote(%q) round-tripped to %q", in, out)
		}
	}
}

func TestWithBinaryHTTPProbe(t *testing.T) {
	app := &models.Application{HealthcheckType: models.HealthcheckHTTP, Port: 8080, HealthcheckHTTPPath: "/healthz"}
	orig := []docker.FileEntry{{Path: "/etc/app.conf", Content: "x"}}

	hc := buildHealthcheck(app)
	if files := withBinaryHTTPProbe(app, hc, orig, "x86_64", true); len(files) != 1 || hc.Test[0] != "CMD-SHELL" {
		t.Fatal("read-only rootfs must keep the shell probe")
	}
	hc = buildHealthcheck(app)
	if files := withBinaryHTTPProbe(app, hc, orig, "s390x", false); len(files) != 1 || hc.Test[0] != "CMD-SHELL" {
		t.Fatal("unsupported arch must keep the shell probe")
	}
	cmd := &models.Application{HealthcheckType: models.HealthcheckCommand, HealthcheckCommand: "true"}
	hc = buildHealthcheck(cmd)
	if files := withBinaryHTTPProbe(cmd, hc, orig, "x86_64", false); len(files) != 1 || hc.Test[0] != "CMD-SHELL" {
		t.Fatal("command healthchecks are left alone")
	}

	if _, ok := healthprobe.Binary("x86_64"); !ok {
		t.Skip("probe binaries not built (make build-probe)")
	}
	hc = buildHealthcheck(app)
	files := withBinaryHTTPProbe(app, hc, orig, "x86_64", false)
	want := []string{"CMD", healthprobe.Path, "8080", "/healthz"}
	if strings.Join(hc.Test, " ") != strings.Join(want, " ") {
		t.Fatalf("test = %v, want %v", hc.Test, want)
	}
	if len(files) != 2 || files[1].Path != healthprobe.Path || files[1].Mode != "0755" || len(files[1].Content) == 0 {
		t.Fatalf("probe file not appended: %+v", files[1:])
	}
	if len(orig) != 1 {
		t.Fatal("caller's file slice must not be modified")
	}
}
