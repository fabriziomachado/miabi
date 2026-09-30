// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package enterprise

import (
	"slices"
	"testing"
)

// ee_features gates capabilities meant for every paid customer, so each tier must grant it.
func TestEEFeaturesEntitlement(t *testing.T) {
	if !IsKnownFlag(FlagEEFeatures) {
		t.Fatalf("%s is missing from AllFlags, so no license can grant it", FlagEEFeatures)
	}
	for _, tier := range Tiers {
		if !slices.Contains(tier.Flags, FlagEEFeatures) {
			t.Errorf("tier %q does not grant %s", tier.Name, FlagEEFeatures)
		}
	}
}
