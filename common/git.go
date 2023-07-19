package common

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5/config"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/go-git/go-git/v5"
)

const (
	defaultHost   = "github.com"
	defaultRemote = "origin"
	gitRepoSuffix = ".git"
)

type cloneRef struct {
	url  string
	base string
}

func processUrl(repoUrl string) (cloneRef, error) {
	var ref cloneRef
	components := strings.Split(repoUrl, "/")
	numComponents := len(components)
	if numComponents == 0 {
		return ref, fmt.Errorf("unexpected repo URL: %s", repoUrl)
	}

	repoBase := components[numComponents-1]
	if strings.HasSuffix(repoBase, gitRepoSuffix) {
		repoBase = repoBase[:len(repoBase)-len(gitRepoSuffix)]
	}
	ref.base = repoBase

	// Doesn't expect git@<host>:<user>/<repo> URLs.
	parsed, err := url.Parse(repoUrl)
	if err != nil {
		return ref, err
	}

	if parsed.Scheme != "" && parsed.Host != "" {
		if !strings.HasSuffix(repoUrl, gitRepoSuffix) {
			repoUrl += gitRepoSuffix
		}
		ref.url = repoUrl
		return ref, nil
	}

	// Assume given URL expects an SSH git clone from this point on.
	if numComponents == 2 {
		repoUrl = fmt.Sprintf("%s:%s", defaultHost, repoUrl)
	} else {
		// URL has host part, change first slash to colon.
		repoUrl = strings.Replace(repoUrl, "/", ":", 1)
	}
	ref.url = fmt.Sprintf("git@%s.git", repoUrl)

	return ref, nil
}

func update(cloneDir string) error {
	r, err := git.PlainOpen(cloneDir)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{RemoteName: defaultRemote})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	return err
}

func clone(repo entity.Repo, repoUrl cloneRef, cloneDir string) error {
	_, err := os.Stat(cloneDir)
	if err == nil {
		return update(cloneDir)
	} else if !os.IsNotExist(err) {
		return err
	}

	opt := git.CloneOptions{
		URL: repoUrl.url,
	}
	if repo.Submodule {
		opt.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	r, err := git.PlainClone(cloneDir, false, &opt)
	if err != nil {
		return fmt.Errorf("error cloning repo %s to %s: %v", repo.Name, cloneDir, err)
	}

	for remote, remoteUrl := range repo.Remotes {
		remoteExpanded, remoteErr := processUrl(remoteUrl)
		_, remoteErr = r.CreateRemote(&config.RemoteConfig{
			Name: remote,
			URLs: []string{remoteExpanded.url},
		})
		if remoteErr != nil {
			return remoteErr
		}

		internal.Log.Debugf("Fetching remote %s from %s for repo %s", remote, remoteUrl, repo.Name)
		remoteErr = r.Fetch(&git.FetchOptions{
			RemoteName: remote,
		})
		if remoteErr != nil {
			return remoteErr
		}
	}

	return nil
}

func CloneTo(repo entity.Repo, path string) error {
	repoUrl, err := processUrl(repo.Name)
	if err != nil {
		return err
	}

	return clone(repo, repoUrl, path)
}

func CloneUnderPath(repo entity.Repo, dir string) error {
	repoUrl, err := processUrl(repo.Name)
	if err != nil {
		return err
	}

	cloneDir := internal.ExpandUser(path.Join(dir, repoUrl.base))
	_, err = os.Stat(cloneDir)
	if err == nil {
		return nil
	}

	return clone(repo, repoUrl, cloneDir)
}
