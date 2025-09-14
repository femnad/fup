package entity

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/femnad/fup/internal"
)

const (
	defaultHost    = "github.com"
	defaultRemote  = "origin"
	gitClonePrefix = "git@"
	gitRepoSuffix  = ".git"
)

var (
	onePasswordSocket = internal.ExpandUser("~/.1password/agent.sock")
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
	repoBase = strings.TrimSuffix(repoBase, gitRepoSuffix)
	ref.base = repoBase

	if strings.HasPrefix(repoUrl, gitClonePrefix) {
		ref.url = repoUrl
		return ref, nil
	}

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

func update(repo Repo, cloneDir string) error {
	if !repo.Update {
		return nil
	}

	r, err := git.PlainOpen(cloneDir)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	slog.Info("Updating repo", "name", repo.Name)

	err = w.Pull(&git.PullOptions{RemoteName: defaultRemote})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	return err
}

func getRef(repo Repo) string {
	if repo.Branch != "" {
		return fmt.Sprintf("refs/head/%s", repo.Branch)
	} else if repo.Tag != "" {
		return fmt.Sprintf("refs/tags/%s", repo.Tag)
	}

	return ""
}

func clone(repo Repo, repoUrl cloneRef, cloneDir string) error {
	_, err := os.Stat(cloneDir)
	if err == nil {
		return update(repo, cloneDir)
	} else if !os.IsNotExist(err) {
		return err
	}

	opt := git.CloneOptions{
		URL: repoUrl.url,
	}
	if repo.Submodule {
		opt.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}
	ref := getRef(repo)
	if ref != "" {
		opt.ReferenceName = plumbing.ReferenceName(ref)
	}

	slog.Info("Cloning repo", "name", repo.Name)

	_, err = os.Stat(onePasswordSocket)
	if err == nil {
		err = os.Setenv("SSH_AUTH_SOCK", onePasswordSocket)
		if err != nil {
			return err
		}
	}

	r, err := git.PlainClone(cloneDir, false, &opt)
	if err != nil {
		return fmt.Errorf("error cloning repo %s to %s: %v", repo.Name, cloneDir, err)
	}

	for remote, remoteUrl := range repo.Remotes {
		remoteExpanded, remoteErr := processUrl(remoteUrl)
		if remoteErr != nil {
			return remoteErr
		}

		_, remoteErr = r.CreateRemote(&config.RemoteConfig{
			Name: remote,
			URLs: []string{remoteExpanded.url},
		})
		if remoteErr != nil {
			return remoteErr
		}

		slog.Debug("Fetching remote", "remote", remote, "url", remoteUrl, "repo", repo.Name)
		remoteErr = r.Fetch(&git.FetchOptions{
			RemoteName: remote,
		})
		if remoteErr != nil {
			return remoteErr
		}
	}

	return nil
}

func CloneUnderPath(repo Repo, dir string, cloneEnv map[string]string) error {
	repoUrl, err := processUrl(repo.Name)
	if err != nil {
		return err
	}

	cloneDir := internal.ExpandUser(path.Join(dir, repoUrl.base))
	_, err = os.Stat(cloneDir)
	if err == nil {
		return nil
	}

	modifiedEnv := make(map[string]string)
	newEnv := make(map[string]bool)
	for k, v := range cloneEnv {
		v = internal.ExpandUser(v)
		envVal, ok := os.LookupEnv(k)
		if ok {
			modifiedEnv[k] = envVal
		} else {
			newEnv[k] = true
		}
		err = os.Setenv(k, v)
		if err != nil {
			return err
		}
	}

	cloneErr := clone(repo, repoUrl, cloneDir)

	for k := range newEnv {
		err = os.Unsetenv(k)
		if err != nil {
			return errors.Join(cloneErr, err)
		}
	}

	for k, v := range modifiedEnv {
		err = os.Setenv(k, v)
		if err != nil {
			return errors.Join(cloneErr, err)
		}
	}

	return cloneErr
}
