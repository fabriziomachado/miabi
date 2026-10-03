// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

package gitops

import (
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// cloneOf creates a repository on defaultBranch plus the given extra branches, one commit each, and
// returns a fresh clone of it with the commit every branch points at.
func cloneOf(t *testing.T, defaultBranch string, extra ...string) (*git.Repository, map[string]plumbing.Hash) {
	t.Helper()
	dir := t.TempDir()
	origin, err := git.PlainInitWithOptions(dir, &git.PlainInitOptions{
		InitOptions: git.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName(defaultBranch)},
	})
	if err != nil {
		t.Fatal(err)
	}
	wt, _ := origin.Worktree()
	commit := func(msg string) plumbing.Hash {
		h, err := wt.Commit(msg, &git.CommitOptions{AllowEmptyCommits: true,
			Author: &object.Signature{Name: "t", Email: "t@example.com", When: time.Now()}})
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	heads := map[string]plumbing.Hash{defaultBranch: commit("init")}
	for _, b := range extra {
		if err := wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(b), Create: true, Hash: heads[defaultBranch]}); err != nil {
			t.Fatal(err)
		}
		heads[b] = commit("on " + b)
	}
	if err := wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(defaultBranch)}); err != nil {
		t.Fatal(err)
	}
	repo, err := git.PlainClone(t.TempDir(), false, &git.CloneOptions{URL: dir})
	if err != nil {
		t.Fatal(err)
	}
	return repo, heads
}

func headOf(t *testing.T, repo *git.Repository) plumbing.Hash {
	t.Helper()
	h, err := repo.Head()
	if err != nil {
		t.Fatal(err)
	}
	return h.Hash()
}

// #451: syncing a branch other than the default failed with "reference not found".
func TestCheckoutRefNonDefaultBranch(t *testing.T) {
	repo, heads := cloneOf(t, "main", "develop")
	if err := checkoutRef(repo, "develop"); err != nil {
		t.Fatal(err)
	}
	if got := headOf(t, repo); got != heads["develop"] {
		t.Errorf("HEAD = %s, want develop at %s", got, heads["develop"])
	}
}

// A source saved without a ref is stored as "main"; on a master-only repository it keeps syncing the
// default branch, as it always has.
func TestCheckoutRefMainFallsBackWithoutMain(t *testing.T) {
	repo, heads := cloneOf(t, "master")
	if err := checkoutRef(repo, "main"); err != nil {
		t.Fatal(err)
	}
	if got := headOf(t, repo); got != heads["master"] {
		t.Errorf("HEAD = %s, want the default branch at %s", got, heads["master"])
	}
}

// ...but where "main" does exist beside the default, it is the branch synced, not the default.
func TestCheckoutRefMainResolvesWhenPresent(t *testing.T) {
	repo, heads := cloneOf(t, "master", "main")
	if err := checkoutRef(repo, "main"); err != nil {
		t.Fatal(err)
	}
	if got := headOf(t, repo); got != heads["main"] {
		t.Errorf("HEAD = %s, want main at %s", got, heads["main"])
	}
}

func TestCheckoutRefUnknownFails(t *testing.T) {
	repo, _ := cloneOf(t, "main")
	if err := checkoutRef(repo, "release-9"); err == nil {
		t.Error("an unknown ref must fail")
	}
}
