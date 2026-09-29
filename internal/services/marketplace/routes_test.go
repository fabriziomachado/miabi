// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package marketplace

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/miabi-io/miabi/internal/models"
	"github.com/miabi-io/miabi/internal/services/marketplace/manifest"
	"github.com/miabi-io/miabi/internal/services/route"
)

type fakeRoutes struct {
	domains map[string]models.Domain // host -> covering domain
	taken   map[string]bool          // route names already used
	created []route.Input
}

func (f *fakeRoutes) Create(_ context.Context, _ uint, in route.Input) (*models.Route, error) {
	if f.taken[in.Name] {
		return nil, route.ErrNameTaken
	}
	if _, ok := f.domains[in.Hosts[0]]; !ok {
		return nil, fmt.Errorf("%w: %s", route.ErrDomainNotRegistered, in.Hosts[0])
	}
	f.created = append(f.created, in)
	return &models.Route{Name: in.Name, Hosts: in.Hosts, ApplicationID: in.ApplicationID}, nil
}

func (f *fakeRoutes) CoveringDomain(_ uint, host string) (*models.Domain, error) {
	if d, ok := f.domains[host]; ok {
		return &d, nil
	}
	return nil, nil
}

func routeTestManifest(t *testing.T) *manifest.Manifest {
	t.Helper()
	m, err := manifest.Parse([]byte(`
apiVersion: miabi.io/v1
kind: Template
metadata: {name: status, displayName: Status, version: "1.0.0"}
inputs:
  - {key: url, label: URL, type: string}
applications:
  - name: web
    image: nginx
    ports:
      - {container: 3000, scheme: http}
routes:
  - app: web
    host: "{{ .inputs.url }}"
`))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func runRoutes(t *testing.T, f *fakeRoutes, url string) ([]*models.Route, []string) {
	t.Helper()
	m := routeTestManifest(t)
	apps := []*models.Application{{ID: 7, DisplayName: "Status"}}
	r := manifest.NewRenderer(manifest.Context{Inputs: map[string]string{"url": url}})
	s := &Service{routes: f}
	return s.createRoutes(context.Background(), 1, m, apps, r, "Status", nil)
}

func TestCreateRoutesVerifiedDomain(t *testing.T) {
	f := &fakeRoutes{domains: map[string]models.Domain{"status.example.com": {Name: "example.com", Verified: true}}}
	routes, warnings := runRoutes(t, f, "https://status.example.com/")
	if len(routes) != 1 || len(warnings) != 0 {
		t.Fatalf("routes=%d warnings=%v", len(routes), warnings)
	}
	in := f.created[0]
	if in.Hosts[0] != "status.example.com" || in.TargetPort != 3000 || in.ApplicationID != 7 || in.Name != "status-web" {
		t.Fatalf("route input = %+v", in)
	}
}

func TestCreateRoutesUnverifiedDomainStillCreates(t *testing.T) {
	f := &fakeRoutes{domains: map[string]models.Domain{"status.example.com": {Name: "example.com"}}}
	routes, warnings := runRoutes(t, f, "https://status.example.com")
	if len(routes) != 1 || len(warnings) != 1 || !strings.Contains(warnings[0], "once example.com is verified") {
		t.Fatalf("routes=%d warnings=%v", len(routes), warnings)
	}
}

func TestCreateRoutesSkipsWithoutFailing(t *testing.T) {
	cases := map[string]string{
		"https://status.example.com": "not registered",
		"":                           "no public URL",
		"http://10.0.0.5:3000":       "not a public hostname",
	}
	for url, want := range cases {
		routes, warnings := runRoutes(t, &fakeRoutes{}, url)
		if len(routes) != 0 || len(warnings) != 1 || !strings.Contains(warnings[0], want) {
			t.Errorf("url %q: routes=%d warnings=%v, want a warning containing %q", url, len(routes), warnings, want)
		}
	}
}

func TestCreateRoutesRetriesTakenName(t *testing.T) {
	f := &fakeRoutes{
		domains: map[string]models.Domain{"status.example.com": {Name: "example.com", Verified: true}},
		taken:   map[string]bool{"status-web": true},
	}
	routes, _ := runRoutes(t, f, "https://status.example.com")
	if len(routes) != 1 || f.created[0].Name != "status-web-2" {
		t.Fatalf("created %+v", f.created)
	}
}
