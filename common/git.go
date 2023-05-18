package common

import (
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

type cloneUrl struct {
	url  string
	base string
}

func processUrl(repoUrl string) (cloneUrl, error) {
	var clone cloneUrl
	components := strings.Split(repoUrl, "/")
	numComponents := len(components)
	if numComponents == 0 {
		return clone, fmt.Errorf("unexpected repo URL: %s", repoUrl)
	}

	repoBase := components[numComponents-1]
	if strings.HasSuffix(repoBase, gitRepoSuffix) {
		repoBase = strings.TrimRight(repoBase, gitRepoSuffix)
	}
	clone.base = repoBase

	// Doesn't expect git@<host>:<user>/<repo> URLs.
	parsed, err := url.Parse(repoUrl)
	if err != nil {
		return clone, err
	}

	if parsed.Scheme != "" && parsed.Host != "" {
		if !strings.HasSuffix(repoUrl, gitRepoSuffix) {
			repoUrl += gitRepoSuffix
		}
		clone.url = repoUrl
		return clone, nil
	}

	// Assume given URL expects an SSH git clone from this point on.
	if numComponents == 2 {
		repoUrl = fmt.Sprintf("%s:%s", defaultHost, repoUrl)
	} else {
		// URL has host part, change first slash to colon.
		repoUrl = strings.Replace(repoUrl, "/", ":", 1)
	}
	clone.url = fmt.Sprintf("git@%s.git", repoUrl)

	return clone, nil
}

func updateSubmodule(r git.Repository) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	submodules, err := w.Submodules()
	if err != nil {
		return err
	}

	for _, submodule := range submodules {
		sub, subErr := w.Submodule(submodule.Config().Name)
		if subErr != nil {
			return subErr
		}

		sr, subErr := sub.Repository()
		if subErr != nil {
			return subErr
		}

		sw, subErr := sr.Worktree()
		if subErr != nil {
			return subErr
		}

		subErr = sw.Pull(&git.PullOptions{
			RemoteName: defaultRemote},
		)
		if subErr != nil {
			return subErr
		}
	}

	return nil
}

func CloneRepo(repo entity.Repo, dir string) error {
	repoUrl, err := processUrl(repo.Name)
	if err != nil {
		return err
	}

	cloneDir := internal.ExpandUser(path.Join(dir, repoUrl.base))
	_, err = os.Stat(cloneDir)
	if err == nil {
		return nil
	}

	opt := git.CloneOptions{
		URL: repoUrl.url,
	}
	if repo.Submodule {
		opt.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	r, err := git.PlainClone(cloneDir, false, &opt)
	if err != nil {
		return fmt.Errorf("error cloning repo %s: %v", repo.Name, err)
	}

	if repo.Submodule {
		err = updateSubmodule(*r)
		if err != nil {
			return err
		}
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
