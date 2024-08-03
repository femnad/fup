package entity

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"

	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/remote"
	"github.com/femnad/fup/settings"
)

const (
	githubLatestRelease = "github-latest"
	githubMatchingTag   = "github-tag"
	pypiLatestVersion   = "pypi-latest"
)

type VersionLookupSpec struct {
	ExcludeSuffix []string `yaml:"exclude_suffix"`
	FollowURL     bool     `yaml:"follow_url"`
	GetFirst      bool     `yaml:"get_first"`
	GetLast       bool     `yaml:"get_last"`
	GithubRepo    string   `yaml:"github_repo"`
	MatchRegex    string   `yaml:"match_regex"`
	PostProc      string   `yaml:"post_proc"`
	Query         string   `yaml:"query"`
	Strategy      string   `yaml:"strategy"`
	URL           string   `yaml:"url"`
}

func resolveQuery(spec VersionLookupSpec) (string, error) {
	doc, err := htmlquery.LoadURL(spec.URL)
	if err != nil {
		return "", err
	}

	query := spec.Query
	if spec.GetFirst {
		node, err := htmlquery.Query(doc, query)
		if err != nil {
			return "", err
		}

		if node == nil {
			return "", fmt.Errorf("error looking up version via query %s", query)
		}

		return htmlquery.InnerText(node), nil
	}

	nodes, err := htmlquery.QueryAll(doc, query)
	if err != nil {
		return "", err
	}

	var regex *regexp.Regexp
	if len(spec.ExcludeSuffix) > 0 {
		suffixPattern := fmt.Sprintf("(%s)$", strings.Join(spec.ExcludeSuffix, "|"))
		regex, err = regexp.Compile(suffixPattern)
		if err != nil {
			return "", err
		}
	}

	numNodes := len(nodes)
	for i, node := range nodes {
		if node == nil {
			continue
		}

		if spec.GetLast && i < numNodes-1 {
			continue
		}

		nodeText := htmlquery.InnerText(node)
		if regex != nil && regex.MatchString(nodeText) {
			continue
		}

		return strings.TrimSpace(nodeText), nil
	}

	return "", fmt.Errorf("unable to find matches via query %s on URL %s", query, spec.URL)
}

func getRepo(spec VersionLookupSpec, releaseURL string) (string, error) {
	repo := spec.GithubRepo
	if repo == "" {
		if releaseURL == "" {
			return "", fmt.Errorf("need a release URL for determining GitHub repo without explicit repo config")
		}

		fields := strings.Split(releaseURL, "/")
		// URL should have the format: https://github.com/<principal>/<repo>/...
		if len(fields) < 5 {
			return "", fmt.Errorf("unable to determine GitHub repo from URL %s", releaseURL)
		}
		repo = fmt.Sprintf("%s/%s", fields[3], fields[4])
	}

	return repo, nil
}

func queryFromStrategy(spec VersionLookupSpec, assetURL string, s settings.Settings) (string, error) {
	resolver := specResolver{useGHClient: s.Internal.GhAvailable}
	strategies := map[string]func(VersionLookupSpec, string) (string, error){
		githubLatestRelease: resolver.githubStable,
		githubMatchingTag:   resolver.gitHubFirstMatchingTag,
		pypiLatestVersion:   resolver.pypiLatestVersion,
	}

	fn, ok := strategies[spec.Strategy]
	if !ok {
		return "", fmt.Errorf("no such strategy %s", spec.Strategy)
	}

	return fn(spec, assetURL)
}

func versionFromSpec(spec VersionLookupSpec, assetURL string, s settings.Settings) (text string, err error) {
	var version string
	if spec.Strategy != "" {
		version, err = queryFromStrategy(spec, assetURL, s)
		if err != nil {
			return "", err
		}
		return version, nil
	}

	if spec.Query != "" {
		text, err = resolveQuery(spec)
		if err != nil {
			return "", err
		}
	} else if spec.FollowURL {
		text, err = remote.FollowRedirects(spec.URL)
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("version lookup requires either a query or follow_url to be set")
	}

	return
}

func lookupVersion(spec VersionLookupSpec, assetURL string, s settings.Settings) (version string, err error) {
	version, err = versionFromSpec(spec, assetURL, s)
	if err != nil {
		return "", err
	}

	if spec.PostProc != "" {
		version, err = internal.RunTemplateFn(version, spec.PostProc)
		if err != nil {
			return "", err
		}
	}

	return version, nil
}
