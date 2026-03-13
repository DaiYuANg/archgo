package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/goyek/goyek/v3"
)

type bumpMode int

const (
	bumpPatch bumpMode = iota
	bumpMinor
	bumpMajor
)

func main() {
	patch := goyek.Define(goyek.Task{
		Name:  "patch",
		Usage: "Auto bump patch version and create local tag",
		Action: func(a *goyek.A) {
			runTagger(a, bumpPatch, false)
		},
	})

	goyek.Define(goyek.Task{
		Name:  "patch-push",
		Usage: "Auto bump patch version and push tag",
		Action: func(a *goyek.A) {
			runTagger(a, bumpPatch, true)
		},
	})

	goyek.Define(goyek.Task{
		Name:  "minor",
		Usage: "Auto bump minor version and create local tag",
		Action: func(a *goyek.A) {
			runTagger(a, bumpMinor, false)
		},
	})

	goyek.Define(goyek.Task{
		Name:  "minor-push",
		Usage: "Auto bump minor version and push tag",
		Action: func(a *goyek.A) {
			runTagger(a, bumpMinor, true)
		},
	})

	goyek.Define(goyek.Task{
		Name:  "major",
		Usage: "Auto bump major version and create local tag",
		Action: func(a *goyek.A) {
			runTagger(a, bumpMajor, false)
		},
	})

	goyek.Define(goyek.Task{
		Name:  "major-push",
		Usage: "Auto bump major version and push tag",
		Action: func(a *goyek.A) {
			runTagger(a, bumpMajor, true)
		},
	})

	goyek.Define(goyek.Task{
		Name:  "help",
		Usage: "Show script usage",
		Action: func(a *goyek.A) {
			_, _ = fmt.Fprintln(a.Output(), "Usage:")
			_, _ = fmt.Fprintln(a.Output(), "  go run ./scripts/tagger [task]")
			_, _ = fmt.Fprintln(a.Output(), "Tasks: patch, patch-push, minor, minor-push, major, major-push, help")
			_, _ = fmt.Fprintln(a.Output(), "Env: TAGGER_REMOTE=origin TAGGER_NAME=auto-tagger TAGGER_EMAIL=ci@local")
		},
	})

	goyek.SetDefault(patch)
	goyek.SetUsage(func() {
		fmt.Fprintln(os.Stderr, "Usage: go run ./scripts/tagger [task]")
		fmt.Fprintln(os.Stderr, "Tasks: patch, patch-push, minor, minor-push, major, major-push, help")
		fmt.Fprintln(os.Stderr, "Env: TAGGER_REMOTE=origin TAGGER_NAME=auto-tagger TAGGER_EMAIL=ci@local")
	})

	goyek.Main(os.Args[1:])
}

func runTagger(a *goyek.A, mode bumpMode, push bool) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		a.Fatal(err)
	}

	latest, err := latestSemverTag(repo)
	if err != nil {
		a.Fatal(err)
	}

	next := bump(latest, mode)
	newTag := "v" + next.String()
	a.Logf("New tag: %s", newTag)

	head, err := repo.Head()
	if err != nil {
		a.Fatal(err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		a.Fatal(err)
	}

	taggerName := getenvDefault("TAGGER_NAME", "auto-tagger")
	taggerEmail := getenvDefault("TAGGER_EMAIL", "ci@local")
	_, err = repo.CreateTag(newTag, commit.Hash, &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  taggerName,
			Email: taggerEmail,
			When:  time.Now(),
		},
		Message: newTag,
	})
	if err != nil {
		a.Fatal(err)
	}
	a.Log("Tag created locally")

	remote := getenvDefault("TAGGER_REMOTE", "origin")
	if !push {
		a.Logf("Push manually: git push %s %s", remote, newTag)
		return
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: remote,
		RefSpecs: []config.RefSpec{
			config.RefSpec("refs/tags/" + newTag + ":refs/tags/" + newTag),
		},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		a.Fatal(err)
	}

	a.Log("Tag pushed")
}

func latestSemverTag(repo *git.Repository) (semver.Version, error) {
	iter, err := repo.Tags()
	if err != nil {
		return semver.Version{}, err
	}

	latest := semver.New(0, 0, 0, "", "")
	found := false

	err = iter.ForEach(func(ref *plumbing.Reference) error {
		tag := ref.Name().Short()
		v, ok := parseSemverTag(tag)
		if !ok {
			return nil
		}
		if !found || v.GreaterThan(latest) {
			latest = v
			found = true
		}
		return nil
	})
	if err != nil {
		return semver.Version{}, err
	}

	if !found {
		return *semver.New(0, 0, 0, "", ""), nil
	}
	return *latest, nil
}

func parseSemverTag(tag string) (*semver.Version, bool) {
	if !strings.HasPrefix(tag, "v") {
		return nil, false
	}

	v, err := semver.StrictNewVersion(strings.TrimPrefix(tag, "v"))
	if err != nil {
		return nil, false
	}
	// Keep compatibility with previous behavior: only stable vX.Y.Z tags.
	if v.Prerelease() != "" || v.Metadata() != "" {
		return nil, false
	}
	return v, true
}

func bump(v semver.Version, mode bumpMode) semver.Version {
	switch mode {
	case bumpMajor:
		return v.IncMajor()
	case bumpMinor:
		return v.IncMinor()
	default:
		return v.IncPatch()
	}
}

func getenvDefault(key, fallback string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val
}
