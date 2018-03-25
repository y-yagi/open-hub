package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func msg(err error, errStream io.Writer) int {
	if err != nil {
		fmt.Fprintf(errStream, "%v\n", err)
		return 1
	}
	return 0
}

func usage(args []string, errStream io.Writer) {
	fmt.Fprintf(errStream, "usage: %s LogHash\n", args[0])
}

func getFullHash(hash string) string {
	// TODO: If `go-git` will support rev-parse, use it.
	out, err := exec.Command("git", "rev-parse", hash).Output()
	if err != nil {
		return hash
	}

	return strings.TrimSpace(string(out))
}

func getRepoURL(r *git.Repository) (string, error) {
	list, err := r.Remotes()
	if err != nil {
		return "", err
	}

	remote := list[0]
	for _, r := range list[1:] {
		if r.Config().Name == "upstream" {
			remote = r
		}
	}

	url := remote.Config().URLs[0]
	url = strings.TrimRight(url, ".git")
	if strings.HasPrefix(url, "https://") {
		return url, nil
	}

	if strings.HasPrefix(url, "git@github") {
		return strings.Replace(url, "git@github.com:", "https://github.com/", -1), nil
	}

	return url, nil
}

func retrivePath(r *git.Repository, hash string) (string, error) {
	commit, err := r.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`Merge pull request #(\d+)`)
	match := re.FindStringSubmatch(commit.Message)
	if match == nil {
		return "/commit/" + hash, nil
	}

	return "/pull/" + match[1], nil
}

func openCommand() string {
	command := ""
	os := runtime.GOOS

	if os == "linux" {
		command = "gnome-open"
	} else if os == "darwin" {
		command = "open"
	}

	return command
}

func run(args []string, outStream, errStream io.Writer) int {
	if len(args) < 2 {
		usage(args, errStream)
		os.Exit(1)
	}

	hash := args[1]
	hash = getFullHash(hash)

	wd, err := os.Getwd()
	if err != nil {
		return msg(err, errStream)
	}

	r, err := git.PlainOpen(wd)
	if err != nil {
		return msg(err, errStream)
	}

	url, err := getRepoURL(r)
	if err != nil {
		return msg(err, errStream)
	}

	path, err := retrivePath(r, hash)
	if err != nil {
		return msg(err, errStream)
	}

	url = url + path
	if err = exec.Command(openCommand(), url).Run(); err != nil {
		return msg(err, errStream)
	}

	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
