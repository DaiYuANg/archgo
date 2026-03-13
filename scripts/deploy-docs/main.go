package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/goyek/goyek/v3"
	goyekcmd "github.com/goyek/x/cmd"
)

type docsContext struct {
	rootDir      string
	docsDir      string
	hugoCacheDir string
}

func main() {
	ctx, err := newDocsContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	build := goyek.Define(goyek.Task{
		Name:  "build",
		Usage: "Build docs with Hugo",
		Action: func(a *goyek.A) {
			if err := os.MkdirAll(ctx.hugoCacheDir, 0o755); err != nil {
				a.Fatal(err)
			}
			if !goyekcmd.Exec(
				a,
				"go tool hugo --gc --minify",
				goyekcmd.Dir(ctx.docsDir),
				goyekcmd.Env("HUGO_CACHEDIR", ctx.hugoCacheDir),
			) {
				a.FailNow()
			}
			a.Logf("build complete: %s", filepath.Join(ctx.docsDir, "public"))
		},
	})

	goyek.Define(goyek.Task{
		Name:  "sync",
		Usage: "Sync version metadata from git tags",
		Action: func(a *goyek.A) {
			if !goyekcmd.Exec(a, "go run ./scripts", goyekcmd.Dir(ctx.docsDir)) {
				a.FailNow()
			}
		},
	})

	goyek.Define(goyek.Task{
		Name:  "serve",
		Usage: "Run local Hugo server",
		Action: func(a *goyek.A) {
			if err := os.MkdirAll(ctx.hugoCacheDir, 0o755); err != nil {
				a.Fatal(err)
			}
			a.Log("visit: http://127.0.0.1:1313")
			if !goyekcmd.Exec(
				a,
				"go tool hugo server -D --bind 0.0.0.0",
				goyekcmd.Dir(ctx.docsDir),
				goyekcmd.Env("HUGO_CACHEDIR", ctx.hugoCacheDir),
			) {
				a.FailNow()
			}
		},
	})

	goyek.Define(goyek.Task{
		Name:  "deploy",
		Usage: "Build and force-push docs/public to gh-pages (set DOCS_REMOTE / DOCS_BRANCH to override defaults)",
		Deps:  goyek.Deps{build},
		Action: func(a *goyek.A) {
			remote := getenvDefault("DOCS_REMOTE", "origin")
			branch := getenvDefault("DOCS_BRANCH", "gh-pages")

			repoURL, err := commandOutput(ctx.rootDir, "git", "remote", "get-url", remote)
			if err != nil {
				a.Fatalf("cannot resolve remote URL for %q: %v", remote, err)
			}

			tempDir := filepath.Join(ctx.docsDir, ".tmp-public")
			if err = os.RemoveAll(tempDir); err != nil {
				a.Fatal(err)
			}
			if err = os.MkdirAll(tempDir, 0o755); err != nil {
				a.Fatal(err)
			}
			if err = copyDirContents(filepath.Join(ctx.docsDir, "public"), tempDir); err != nil {
				a.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(tempDir, ".nojekyll"), []byte{}, 0o644); err != nil {
				a.Fatal(err)
			}

			execOrFail(a, tempDir, "git", "init")
			execOrFail(a, tempDir, "git", "checkout", "-b", branch)
			execOrFail(a, tempDir, "git", "add", "-A")

			if isCleanGitIndex(tempDir) {
				a.Log("no changes to deploy")
				return
			}

			commitMsg := "docs: deploy " + time.Now().UTC().Format(time.RFC3339)
			execOrFail(a, tempDir, "git", "commit", "-m", commitMsg)
			execOrFail(a, tempDir, "git", "remote", "add", "origin", repoURL)
			execOrFail(a, tempDir, "git", "push", "-f", "origin", branch)
			a.Logf("deployed to %s/%s", remote, branch)
		},
	})

	goyek.Define(goyek.Task{
		Name:  "help",
		Usage: "Show script usage",
		Action: func(a *goyek.A) {
			_, _ = fmt.Fprintln(a.Output(), "Usage:")
			_, _ = fmt.Fprintln(a.Output(), "  go run ./scripts/deploy-docs <task>")
			_, _ = fmt.Fprintln(a.Output(), "Tasks: sync, build, serve, deploy, help")
			_, _ = fmt.Fprintln(a.Output(), "Env (deploy): DOCS_REMOTE=origin DOCS_BRANCH=gh-pages")
		},
	})

	goyek.SetUsage(func() {
		fmt.Fprintln(os.Stderr, "Usage: go run ./scripts/deploy-docs [task]")
		fmt.Fprintln(os.Stderr, "Tasks: sync, build, serve, deploy, help")
		fmt.Fprintln(os.Stderr, "Deploy env: DOCS_REMOTE=origin DOCS_BRANCH=gh-pages")
	})

	goyek.Main(os.Args[1:])
}

func execOrFail(a *goyek.A, dir string, name string, args ...string) {
	cmdLine := name + " " + strings.Join(args, " ")
	if !goyekcmd.Exec(a, cmdLine, goyekcmd.Dir(dir)) {
		a.FailNow()
	}
}

func newDocsContext() (*docsContext, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("cannot resolve script path")
	}
	scriptDir := filepath.Dir(thisFile)
	rootDir := filepath.Clean(filepath.Join(scriptDir, "..", ".."))
	docsDir := filepath.Join(rootDir, "docs")

	info, err := os.Stat(docsDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("invalid docs directory: %s", docsDir)
	}

	return &docsContext{
		rootDir:      rootDir,
		docsDir:      docsDir,
		hugoCacheDir: filepath.Join(docsDir, ".cache", "hugo"),
	}, nil
}

func commandOutput(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func isCleanGitIndex(dir string) bool {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = dir
	return cmd.Run() == nil
}

func copyDirContents(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())
		if entry.IsDir() {
			if err = copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		if err = copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func copyDir(srcDir, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}
	return copyDirContents(srcDir, dstDir)
}

func copyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	_, err = io.Copy(dst, src)
	return err
}

func getenvDefault(key, fallback string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val
}
