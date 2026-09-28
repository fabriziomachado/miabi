// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package mwcatalog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/miabi-io/miabi/internal/services/crypto"
)

type versionedKeyring struct{ active *int }

func (k versionedKeyring) dek(ver int) []byte { return bytes.Repeat([]byte{byte(ver)}, 32) }
func (k versionedKeyring) ActiveDEK(uint) (int, []byte, error) {
	return *k.active, k.dek(*k.active), nil
}
func (k versionedKeyring) DEK(_ uint, ver int) ([]byte, error) { return k.dek(ver), nil }

func TestReencryptSecretsMovesToActiveVersion(t *testing.T) {
	active := 1
	crypto.Init("test-master-key-for-mwcatalog")
	crypto.SetKeyring(versionedKeyring{active: &active})
	t.Cleanup(func() { crypto.SetKeyring(nil) })

	enc, err := EncryptSecrets("basicAuth", 5, map[string]any{
		"users": []any{
			map[string]any{"username": "jude", "password": "s3cret"},
			map[string]any{"username": "legacy", "password": "plain"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// A rule saved before secrets were encrypted still holds plaintext.
	enc["users"].([]any)[1].(map[string]any)["password"] = "plain"

	active = 2
	out, changed, err := ReencryptSecrets("basicAuth", 5, enc)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("changed = false after a key version bump")
	}
	users := out["users"].([]any)
	if pw := users[0].(map[string]any)["password"].(string); !strings.HasPrefix(pw, "e2:w:5:2:") {
		t.Fatalf("password not moved to v2: %q", pw)
	}
	if pw := users[1].(map[string]any)["password"]; pw != "plain" {
		t.Fatalf("legacy plaintext was rewritten: %q", pw)
	}
	dec, err := DecryptSecrets("basicAuth", out)
	if err != nil {
		t.Fatal(err)
	}
	if pw := dec["users"].([]any)[0].(map[string]any)["password"]; pw != "s3cret" {
		t.Fatalf("decrypted = %v, want s3cret", pw)
	}

	if _, changed, err := ReencryptSecrets("basicAuth", 5, out); err != nil || changed {
		t.Fatalf("second pass: changed=%v err=%v, want no-op", changed, err)
	}
}
