package entity

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"

	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/remote"
)

const (
	githubLatestRelease = "github-latest"
)

var (
	strategies = map[string]func(VersionLookupSpec, string) (string, error){
		githubLatestRelease: githubStableSpec,
	}
)

type githubReleaseResp struct {
	TagName string `json:"tag_name"`
}

type VersionLookupSpec struct {
	ExcludeSuffix []string `yaml:"exclude_suffix"`
	FollowURL     bool     `yaml:"follow_url"`
	GetFirst      bool     `yaml:"get_first"`
	GithubRepo    string   `yaml:"github_repo"`
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

	for _, node := range nodes {
		if node == nil {
			continue
		}

		nodeText := htmlquery.InnerText(node)
		if regex != nil && regex.MatchString(nodeText) {
			continue
		}

		return nodeText, nil
	}

	return "", fmt.Errorf("unable to find matches via query %s on URL %s", query, spec.URL)
}

func githubStableSpec(spec VersionLookupSpec, archiveURL string) (string, error) {
	repo := spec.GithubRepo
	if repo == "" {
		fields := strings.Split(archiveURL, "/")
		// URL should have the format: https://github.com/<principal>/<repo>/...
		if len(fields) < 5 {
			return "", fmt.Errorf("unable to determine GitHub repo from URL %s", archiveURL)
		}
		repo = fmt.Sprintf("%s/%s", fields[3], fields[4])
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := remote.ReadResponseBytes(apiURL)
	if err != nil {
		return "", err
	}

	var releaseResp githubReleaseResp
	err = json.Unmarshal(resp, &releaseResp)
	if err != nil {
		return "", err
	}

	latestTag := releaseResp.TagName
	if latestTag == "" {
		return "", fmt.Errorf("querying latest %s release returned an empty tag", repo)
	}

	return latestTag, nil
}

func queryFromStrategy(spec VersionLookupSpec, archiveURL string) (string, error) {
	fn, ok := strategies[spec.Strategy]
	if !ok {
		return "", fmt.Errorf("no such strategy %s", spec.Strategy)
	}

	return fn(spec, archiveURL)
}

func versionFromSpec(spec VersionLookupSpec, archiveURL string) (text string, err error) {
	var version string
	if spec.Strategy != "" {
		version, err = queryFromStrategy(spec, archiveURL)
		if err != nil {
			return "", err
		}
		return version, nil
	}

	if spec.Query != "" {
		text, err = resolveQuery(spec)
		if err != nil {
			if err != nil {
				return "", err
			}
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

func lookupVersion(spec VersionLookupSpec, archiveURL string) (version string, err error) {
	version, err = versionFromSpec(spec, archiveURL)
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
