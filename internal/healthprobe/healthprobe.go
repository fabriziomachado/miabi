// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package healthprobe embeds the static HTTP probe (cmd/healthprobe) that the deploy
// worker copies into app containers. bin/ is a build artifact (`make build-probe`):
// only a .gitkeep is committed, so a build without it falls back to a shell probe.
package healthprobe

import (
	"embed"
	"strings"
)

// Path is where the probe is placed inside a container.
const Path = "/.miabi/healthprobe"

//go:embed all:bin
var bins embed.FS

// Binary returns the probe for a Docker-reported architecture, or false when
// this build doesn't carry one.
func Binary(arch string) ([]byte, bool) {
	goarch, ok := goArch(arch)
	if !ok {
		return nil, false
	}
	b, err := bins.ReadFile("bin/healthprobe-linux-" + goarch)
	if err != nil || len(b) == 0 {
		return nil, false
	}
	return b, true
}

// Bundled reports whether this build carries a probe for any architecture.
func Bundled() bool {
	for _, arch := range []string{"amd64", "arm64"} {
		if _, ok := Binary(arch); ok {
			return true
		}
	}
	return false
}

func goArch(arch string) (string, bool) {
	switch strings.ToLower(arch) {
	case "x86_64", "amd64":
		return "amd64", true
	case "aarch64", "arm64":
		return "arm64", true
	}
	return "", false
}
