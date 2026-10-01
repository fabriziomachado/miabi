// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package healthprobe

import "testing"

func TestGoArch(t *testing.T) {
	for in, want := range map[string]string{"x86_64": "amd64", "amd64": "amd64", "aarch64": "arm64", "ARM64": "arm64"} {
		if got, ok := goArch(in); !ok || got != want {
			t.Errorf("goArch(%q) = %q, %v", in, got, ok)
		}
	}
	if _, ok := goArch("s390x"); ok {
		t.Error("unsupported arch should report false")
	}
	if _, ok := Binary("s390x"); ok {
		t.Error("Binary should be absent for an unsupported arch")
	}
}
