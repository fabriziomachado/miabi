// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import (
	"errors"
	"fmt"
	"strings"

	"github.com/miabi-io/miabi/internal/models"
)

// Kinds of an app's database link.
const (
	LinkKindDatabase = "database" // a logical database the app owns
	LinkKindInstance = "instance" // a whole instance (Redis), shareable across apps
)

var (
	// ErrLinkLogicalDatabase refuses an instance link on an engine whose apps
	// connect through a logical database with its own scoped user.
	ErrLinkLogicalDatabase = errors.New("this engine is linked through one of its databases, not the instance")
	ErrInvalidEnvMap       = errors.New("invalid env mapping")
	ErrEnvConflict         = errors.New("env vars already injected by another linked database")
)

// EnvLink is how a link's connection lands in the app env: the prefix and
// mapping the user chose, and the var names they resolve to.
type EnvLink struct {
	Prefix string
	Map    map[string]string
	Vars   []string
}

// ResolveEnvLink validates a prefix and mapping for an engine and refuses var
// names another of the app's links already injects. skipDB/skipInst exclude the
// link being (re)attached.
func (s *Service) ResolveEnvLink(workspaceID, appID uint, engine models.DBEngine, prefix string, m map[string]string, skipDB, skipInst uint) (EnvLink, error) {
	norm, err := NormalizeEnvMap(engine, prefix, m)
	if err != nil {
		return EnvLink{}, fmt.Errorf("%w: %s", ErrInvalidEnvMap, err.Error())
	}
	keys := EnvKeys(engine, prefix, norm)
	linked, err := s.ListByApp(workspaceID, appID)
	if err != nil {
		return EnvLink{}, err
	}
	var taken []string
	for _, l := range linked {
		if (l.Kind == LinkKindDatabase && l.ID == skipDB) || (l.Kind == LinkKindInstance && l.ID == skipInst) {
			continue
		}
		taken = append(taken, l.EnvVars...)
	}
	if c := envConflicts(keys, taken); len(c) > 0 {
		return EnvLink{}, fmt.Errorf("%w: %s; set a prefix or map them to other names", ErrEnvConflict, strings.Join(c, ", "))
	}
	return EnvLink{Prefix: prefix, Map: norm, Vars: keys}, nil
}

// LinkInstance links a whole instance to an app, or updates an existing link's
// env. Only engines without per-app logical databases (Redis) link this way.
func (s *Service) LinkInstance(workspaceID, instanceID, appID uint, env EnvLink) (*models.DatabaseInstanceLink, error) {
	inst, err := s.repo.FindInWorkspace(workspaceID, instanceID)
	if err != nil {
		return nil, ErrNotFound
	}
	if models.EngineUsesLogicalDatabaseRecord(inst.Engine) {
		return nil, ErrLinkLogicalDatabase
	}
	l, err := s.repo.FindInstanceLink(workspaceID, instanceID, appID)
	if err != nil {
		l = &models.DatabaseInstanceLink{WorkspaceID: workspaceID, InstanceID: instanceID, ApplicationID: appID}
	}
	l.EnvPrefix, l.EnvMap, l.EnvVars = env.Prefix, env.Map, env.Vars
	if err := s.repo.SaveInstanceLink(l); err != nil {
		return nil, err
	}
	return l, nil
}

// UnlinkInstance removes an app's link to an instance and returns it, so the
// caller can clean up the vars it injected.
func (s *Service) UnlinkInstance(workspaceID, instanceID, appID uint) (*models.DatabaseInstanceLink, error) {
	l, err := s.repo.FindInstanceLink(workspaceID, instanceID, appID)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.repo.DeleteInstanceLink(l.ID); err != nil {
		return nil, err
	}
	return l, nil
}

// InstanceConnectionForApp reveals a linked instance's connection to an app
// that holds a link to it.
func (s *Service) InstanceConnectionForApp(workspaceID, appID, instanceID uint) (ConnectionInfo, error) {
	if _, err := s.repo.FindInstanceLink(workspaceID, instanceID, appID); err != nil {
		return ConnectionInfo{}, ErrNotFound
	}
	inst, err := s.repo.FindInWorkspace(workspaceID, instanceID)
	if err != nil {
		return ConnectionInfo{}, ErrNotFound
	}
	return s.InstanceConnection(inst)
}

// ListInstanceLinks returns every app link to an instance.
func (s *Service) ListInstanceLinks(instanceID uint) ([]models.DatabaseInstanceLink, error) {
	return s.repo.ListInstanceLinks(instanceID)
}

// InstanceLinkForApp loads an app's link to an instance.
func (s *Service) InstanceLinkForApp(workspaceID, instanceID, appID uint) (*models.DatabaseInstanceLink, error) {
	l, err := s.repo.FindInstanceLink(workspaceID, instanceID, appID)
	if err != nil {
		return nil, ErrNotFound
	}
	return l, nil
}
