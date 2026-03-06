package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var semverTag = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

type version struct {
	major int
	minor int
	patch int
}

func (v version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func (v version) Less(other version) bool {
	if v.major != other.major {
		return v.major < other.major
	}
	if v.minor != other.minor {
		return v.minor < other.minor
	}
	return v.patch < other.patch
}

type bumpMode int

const (
	bumpPatch bumpMode = iota
	bumpMinor
	bumpMajor
)

func main() {
	push := flag.Bool("push", false, "push the created tag to remote")
	remote := flag.String("remote", "origin", "remote used when --push is enabled")
	minor := flag.Bool("minor", false, "bump minor version")
	major := flag.Bool("major", false, "bump major version")
	patch := flag.Bool("patch", false, "bump patch version (default)")
	taggerName := flag.String("tagger-name", "auto-tagger", "tagger name")
	taggerEmail := flag.String("tagger-email", "ci@local", "tagger email")
	flag.Parse()

	mode := resolveBumpMode(*major, *minor, *patch)

	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Fatal(err)
	}

	latest, err := latestSemverTag(repo)
	if err != nil {
		log.Fatal(err)
	}

	next := bump(latest, mode)
	newTag := next.String()
	fmt.Println("New tag:", newTag)

	head, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		log.Fatal(err)
	}

	_, err = repo.CreateTag(newTag, commit.Hash, &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  *taggerName,
			Email: *taggerEmail,
			When:  time.Now(),
		},
		Message: newTag,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Tag created locally")

	if !*push {
		fmt.Printf("Push manually: git push %s %s\n", *remote, newTag)
		return
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: *remote,
		RefSpecs: []config.RefSpec{
			config.RefSpec("refs/tags/" + newTag + ":refs/tags/" + newTag),
		},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Fatal(err)
	}

	fmt.Println("Tag pushed")
}

func resolveBumpMode(major bool, minor bool, patch bool) bumpMode {
	enabled := 0
	if major {
		enabled++
	}
	if minor {
		enabled++
	}
	if patch {
		enabled++
	}
	if enabled > 1 {
		log.Fatal("only one of --major, --minor, --patch can be set")
	}

	if major {
		return bumpMajor
	}
	if minor {
		return bumpMinor
	}
	return bumpPatch
}

func latestSemverTag(repo *git.Repository) (version, error) {
	iter, err := repo.Tags()
	if err != nil {
		return version{}, err
	}

	latest := version{}
	found := false

	err = iter.ForEach(func(ref *plumbing.Reference) error {
		tag := ref.Name().Short()
		v, ok := parseSemverTag(tag)
		if !ok {
			return nil
		}
		if !found || latest.Less(v) {
			latest = v
			found = true
		}
		return nil
	})
	if err != nil {
		return version{}, err
	}

	if !found {
		return version{major: 0, minor: 0, patch: 0}, nil
	}
	return latest, nil
}

func parseSemverTag(tag string) (version, bool) {
	match := semverTag.FindStringSubmatch(tag)
	if match == nil {
		return version{}, false
	}

	major, err := strconv.Atoi(match[1])
	if err != nil {
		return version{}, false
	}
	minor, err := strconv.Atoi(match[2])
	if err != nil {
		return version{}, false
	}
	patch, err := strconv.Atoi(match[3])
	if err != nil {
		return version{}, false
	}

	return version{
		major: major,
		minor: minor,
		patch: patch,
	}, true
}

func bump(v version, mode bumpMode) version {
	switch mode {
	case bumpMajor:
		return version{major: v.major + 1}
	case bumpMinor:
		return version{major: v.major, minor: v.minor + 1}
	default:
		return version{major: v.major, minor: v.minor, patch: v.patch + 1}
	}
}
