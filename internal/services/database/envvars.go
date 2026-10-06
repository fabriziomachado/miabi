// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"

	"github.com/miabi-io/miabi/internal/models"
)

// Connection fields a linked database is injected into an app as. An env map
// keys on these to rename a field's variable, or blank it to skip the field.
const (
	EnvFieldURL      = "url"
	EnvFieldLegacy   = "database_url" // deprecated alias of url
	EnvFieldHost     = "host"
	EnvFieldPort     = "port"
	EnvFieldName     = "name"
	EnvFieldUser     = "user"
	EnvFieldPassword = "password"
)

// envFieldOrder fixes the order variables are written in.
var envFieldOrder = []string{EnvFieldURL, EnvFieldLegacy, EnvFieldHost, EnvFieldPort, EnvFieldName, EnvFieldUser, EnvFieldPassword}

var envNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// DefaultEnvNames returns the variable each connection field is injected as
// for an engine, before any prefix. Redis has no user or database name, and
// takes REDIS_* so it can sit next to an unprefixed SQL database.
func DefaultEnvNames(engine models.DBEngine) map[string]string {
	if engine == models.DBEngineRedis {
		return map[string]string{
			EnvFieldURL:      "REDIS_URL",
			EnvFieldHost:     "REDIS_HOST",
			EnvFieldPort:     "REDIS_PORT",
			EnvFieldPassword: "REDIS_PASSWORD",
		}
	}
	return map[string]string{
		EnvFieldURL:      "DB_URL",
		EnvFieldLegacy:   "DATABASE_URL",
		EnvFieldHost:     "DB_HOST",
		EnvFieldPort:     "DB_PORT",
		EnvFieldName:     "DB_NAME",
		EnvFieldUser:     "DB_USER",
		EnvFieldPassword: "DB_PASSWORD",
	}
}

// NormalizeEnvMap validates a user-supplied env map for an engine: every key
// must be a field the engine injects, every non-empty value a valid variable
// name, and no two fields may resolve to the same variable.
func NormalizeEnvMap(engine models.DBEngine, prefix string, m map[string]string) (map[string]string, error) {
	if len(m) == 0 {
		return nil, nil
	}
	defaults := DefaultEnvNames(engine)
	out := make(map[string]string, len(m))
	for field, name := range m {
		if _, ok := defaults[field]; !ok {
			return nil, fmt.Errorf("env field %q does not apply to %s", field, engine)
		}
		if name != "" && !envNameRe.MatchString(name) {
			return nil, fmt.Errorf("invalid env var name %q for %s", name, field)
		}
		out[field] = name
	}
	seen := map[string]string{}
	for field, name := range resolveEnvNames(engine, prefix, out) {
		if other, dup := seen[name]; dup {
			return nil, fmt.Errorf("%s and %s both map to %s", other, field, name)
		}
		seen[name] = field
	}
	return out, nil
}

// EnvKeys returns the variable names a link injects, in write order.
func EnvKeys(engine models.DBEngine, prefix string, m map[string]string) []string {
	names := resolveEnvNames(engine, prefix, m)
	out := make([]string, 0, len(names))
	for _, f := range envFieldOrder {
		if n, ok := names[f]; ok {
			out = append(out, n)
		}
	}
	return out
}

// resolveEnvNames maps each field the engine injects to its final variable
// name; skipped fields are absent.
func resolveEnvNames(engine models.DBEngine, prefix string, m map[string]string) map[string]string {
	out := map[string]string{}
	for field, def := range DefaultEnvNames(engine) {
		name, mapped := m[field]
		if !mapped {
			name = def
			if prefix != "" {
				name = prefix + "_" + def
			}
		}
		if name != "" {
			out[field] = name
		}
	}
	return out
}

// EnvVar is one variable a database link writes onto an app.
type EnvVar struct {
	Key    string
	Value  string
	Secret bool
}

// AppEnv builds the variables that inject a database's connection into an app:
// the logical database's scoped connection when d is set, else the instance's
// own (Redis). Password and URL are Vault references when secrets are wired, so
// the app env never holds plaintext; otherwise they are stored as secret vars.
func (s *Service) AppEnv(inst *models.DatabaseInstance, d *models.Database, prefix string, m map[string]string) ([]EnvVar, error) {
	var conn ConnectionInfo
	var err error
	var urlSecret, passSecret string
	if d != nil {
		conn, err = s.DatabaseConnection(inst, d)
		urlSecret, passSecret = URLSecretName(inst, d), PasswordSecretName(inst, d)
	} else {
		conn, err = s.InstanceConnection(inst)
		urlSecret, passSecret = InstanceURLSecretName(inst), InstancePasswordSecretName(inst)
	}
	if err != nil {
		return nil, err
	}
	url := EnvVar{Value: conn.URI, Secret: true}
	pass := EnvVar{Value: conn.Password, Secret: true}
	if s.secrets != nil {
		if conn.URI != "" {
			url = EnvVar{Value: "${{ secrets." + urlSecret + " }}"}
		}
		if conn.Password != "" {
			pass = EnvVar{Value: "${{ secrets." + passSecret + " }}"}
		}
	}
	values := map[string]EnvVar{
		EnvFieldURL:      url,
		EnvFieldLegacy:   url,
		EnvFieldHost:     {Value: conn.Host},
		EnvFieldPort:     {Value: strconv.Itoa(conn.Port)},
		EnvFieldName:     {Value: conn.Database},
		EnvFieldUser:     {Value: conn.Username},
		EnvFieldPassword: pass,
	}
	names := resolveEnvNames(inst.Engine, prefix, m)
	var out []EnvVar
	for _, f := range envFieldOrder {
		name, ok := names[f]
		v := values[f]
		if !ok || v.Value == "" || (f == EnvFieldPort && conn.Port == 0) {
			continue
		}
		v.Key = name
		out = append(out, v)
	}
	return out, nil
}

// InjectedKeys returns the variables a link put on its app: the recorded set,
// or, for links made before keys were recorded, the set its prefix implies.
func InjectedKeys(engine models.DBEngine, prefix string, m map[string]string, recorded []string) []string {
	if len(recorded) > 0 {
		return recorded
	}
	return EnvKeys(engine, prefix, m)
}

// envConflicts returns the keys that are already injected by another link.
func envConflicts(keys, taken []string) []string {
	var out []string
	for _, k := range keys {
		if slices.Contains(taken, k) {
			out = append(out, k)
		}
	}
	return out
}
