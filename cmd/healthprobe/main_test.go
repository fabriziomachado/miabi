// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/redirect":
			w.WriteHeader(http.StatusFound)
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer srv.Close()
	_, port, _ := net.SplitHostPort(srv.Listener.Addr().String())

	cases := map[string]int{"/ok": 0, "/redirect": 0, "/down": 1}
	for path, want := range cases {
		if got := probe([]string{"127.0.0.1"}, port, path, 2*time.Second); got != want {
			t.Errorf("probe %s = %d, want %d", path, got, want)
		}
	}
}

func TestProbeFallsBackToNextHost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer srv.Close()
	_, port, _ := net.SplitHostPort(srv.Listener.Addr().String())

	if got := probe([]string{"127.0.0.2", "127.0.0.1"}, port, "/", time.Second); got != 0 {
		t.Errorf("want fallback to the listening host, got %d", got)
	}
}

func TestProbeNothingListening(t *testing.T) {
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	_, port, _ := net.SplitHostPort(l.Addr().String())
	_ = l.Close()
	if got := probe([]string{"127.0.0.1"}, port, "/", time.Second); got != 1 {
		t.Errorf("closed port should fail, got %d", got)
	}
}

func TestStatusCode(t *testing.T) {
	for line, want := range map[string]int{
		"HTTP/1.1 204 No Content\r\n": 204,
		"HTTP/1.0 503\r\n":            503,
		"garbage":                     0,
		"":                            0,
	} {
		if got := statusCode(line); got != want {
			t.Errorf("statusCode(%q) = %d, want %d", line, got, want)
		}
	}
}
