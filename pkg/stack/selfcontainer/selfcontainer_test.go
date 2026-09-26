// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: Apache-2.0

package selfcontainer

import "testing"

func TestMatch(t *testing.T) {
	full := "3f5d2c1a9b8e7d6c5b4a39281706f5e4d3c2b1a09f8e7d6c5b4a3928170615243"
	short := full[:12]

	cases := []struct {
		name string
		a, b string
		want bool
	}{
		{"full equals full", full, full, true},
		{"short prefix of full", short, full, true},
		{"full has short prefix", full, short, true},
		{"empty self never matches", "", full, false},
		{"empty target never matches", full, "", false},
		{"too short to match", full[:8], full, false},
		{"different ids", full, "a" + full[1:], false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Match(tc.a, tc.b); got != tc.want {
				t.Fatalf("Match(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestDetectEnvOverride(t *testing.T) {
	t.Setenv("MIABI_CONTAINER_ID", "  deadbeefcafe  ")
	if got := Detect(); got != "deadbeefcafe" {
		t.Fatalf("Detect() = %q, want trimmed override", got)
	}
}

// The container ID is the one right after /containers/, not the first 64-hex run on the line: under Docker
// in Docker the data root sits inside a volume whose name comes first.
func TestIDOnMountinfo(t *testing.T) {
	const self = "85a39ae487e95d0846318bc10d15f7979a16fcca1320391650450cef28f8e807"
	const vol = "9a5abee34f0e1e80c26e7ac76ee2571892550ca01e7e840aa3d7c93041b78481"
	cases := map[string]string{
		"5798 5789 252:15 /var/lib/docker/containers/" + self + "/hostname /etc/hostname rw":                           self,
		"5798 5789 252:15 /var/lib/docker/volumes/" + vol + "/_data/containers/" + self + "/hostname /etc/hostname rw": self,
		"5798 5789 252:15 /var/lib/docker/overlay2/" + vol + "/merged / rw":                                            "",
	}
	for line, want := range cases {
		if got := idOn(line, "/containers/"); got != want {
			t.Errorf("idOn(%q) = %q, want %q", line, got, want)
		}
	}
}
