package common

import (
	"fmt"
	"github.com/femnad/fup/internal"
	"github.com/go-git/go-git/v5"
	"net/url"
	"os"
	"path"
	"strings"
)

const (
	defaultHost   = "github.com"
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

func CloneRepo(repo, dir string) error {
	repoUrl, err := processUrl(repo)
	if err != nil {
		return err
	}

	cloneDir := internal.ExpandUser(path.Join(dir, repoUrl.url))
	_, err = os.Stat(cloneDir)
	if err == nil {
		return nil
	}

	opt := git.CloneOptions{
		URL: repoUrl.url,
	}
	_, err = git.PlainClone(cloneDir, false, &opt)
	if err != nil {
		return fmt.Errorf("error cloning repo %s: %v", repo, err)
	}

	return nil
}
