// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package models

import "testing"

func TestHealthcheckGatesDeploy(t *testing.T) {
	cases := map[HealthcheckType]bool{
		HealthcheckHTTP: true, HealthcheckCommand: true,
		HealthcheckImage: false, HealthcheckNone: false, "": false,
	}
	for typ, want := range cases {
		if got := typ.GatesDeploy(); got != want {
			t.Errorf("%q.GatesDeploy() = %v, want %v", typ, got, want)
		}
	}
	if !ValidHealthcheckType(HealthcheckImage) {
		t.Error("image must be a valid healthcheck type")
	}
}
