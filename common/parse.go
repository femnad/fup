package common

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

func NameFromRepo(repo string) (string, error) {
	u, err := url.Parse(repo)
	if err != nil {
		return "", err
	}

	_, repoName := path.Split(u.Path)
	split := strings.Split(repoName, ".")
	if len(split) != 2 {
		return "", fmt.Errorf("unexpected repo name %s", repoName)
	}

	return split[0], nil
}
