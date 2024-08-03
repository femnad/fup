package entity

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/cli/go-gh/v2/pkg/api"

	"github.com/femnad/fup/remote"
)

const (
	apiBase = "https://api.github.com"
)

type specResolver struct {
	useGHClient bool
}

type githubReleaseResp struct {
	TagName string `json:"tag_name"`
}

type githubTagResp struct {
	Name string `json:"name"`
}

type ghRequestSpec struct {
	lookupSpec  VersionLookupSpec
	releaseURL  string
	apiPathSpec string
	useGHClient bool
}

func githubRequest[t interface{}](spec ghRequestSpec) (respEntity t, err error) {
	repo, err := getRepo(spec.lookupSpec, spec.releaseURL)
	if err != nil {
		return respEntity, err
	}
	apiPath := fmt.Sprintf(spec.apiPathSpec, repo)

	if spec.useGHClient {
		var client *api.RESTClient
		client, err = api.DefaultRESTClient()
		if err != nil {
			return
		}

		err = client.Get(apiPath, &respEntity)
		return
	}

	url := fmt.Sprintf("%s/%s", apiBase, apiPath)
	var resp []byte
	resp, err = remote.ReadResponseBytes(url)
	if err != nil {
		return
	}

	err = json.Unmarshal(resp, &respEntity)
	return

}

func (s specResolver) githubStable(spec VersionLookupSpec, releaseURL string) (string, error) {
	release, err := githubRequest[githubReleaseResp](ghRequestSpec{
		lookupSpec:  spec,
		releaseURL:  releaseURL,
		apiPathSpec: "repos/%s/releases/latest",
		useGHClient: s.useGHClient})
	if err != nil {
		return "", err
	}

	return release.TagName, nil
}

func (s specResolver) gitHubFirstMatchingTag(spec VersionLookupSpec, releaseURL string) (out string, err error) {
	var regex *regexp.Regexp
	if spec.MatchRegex != "" {
		regex, err = regexp.Compile(spec.MatchRegex)
		if err != nil {
			return
		}
	}

	tags, err := githubRequest[[]githubTagResp](ghRequestSpec{
		lookupSpec:  spec,
		releaseURL:  releaseURL,
		apiPathSpec: "repos/%s/releases/latest",
		useGHClient: s.useGHClient})
	if err != nil {
		return
	}

	for _, tag := range tags {
		if regex != nil && !regex.MatchString(tag.Name) {
			continue
		}

		return tag.Name, nil
	}

	return "", fmt.Errorf("error finding matching tag for spec %+v", spec)
}

func (s specResolver) pypiLatestVersion(spec VersionLookupSpec, pkgName string) (string, error) {
	packageURL := fmt.Sprintf("https://pypi.org/project/%s/", pkgName)
	lookupSpec := VersionLookupSpec{
		Query: "//h1[@class='package-header__name']",
		URL:   packageURL,
	}
	return resolveQuery(lookupSpec)
}
