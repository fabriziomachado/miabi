// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package wsbackup

import (
	"testing"

	"github.com/miabi-io/miabi/internal/models"
	"github.com/miabi-io/miabi/internal/wsbundle"
)

// A backup taken before "none" disabled the image's own check must restore to what it ran with.
func TestRestoredHealthcheckType(t *testing.T) {
	cases := []struct {
		typ    string
		schema int
		want   models.HealthcheckType
	}{
		{"none", 2, models.HealthcheckImage},
		{"none", wsbundle.HealthcheckNoneDisables, models.HealthcheckNone},
		{"http", 2, models.HealthcheckHTTP},
		{"image", wsbundle.StateSchema, models.HealthcheckImage},
	}
	for _, c := range cases {
		if got := restoredHealthcheckType(c.typ, c.schema); got != c.want {
			t.Errorf("%q at schema %d = %q, want %q", c.typ, c.schema, got, c.want)
		}
	}
}
