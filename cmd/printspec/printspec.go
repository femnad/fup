package printspec

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/femnad/fup/cmd/base"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"gopkg.in/yaml.v3"
)

var (
	// Negative matches.
	appleRegex       = regexp.MustCompile("(?i)apple")
	darwinRegex      = regexp.MustCompile("(?i)darwin")
	freeBSRegex      = regexp.MustCompile("(?i)freebsd")
	packageFileRegex = regexp.MustCompile("(?i)\\.(apk|deb|rpm)$")
	windowsRegex     = regexp.MustCompile("(?i)windows")
	// Positive matches.
	architectureRegex = regexp.MustCompile("(?i)x86[_-]64")
	linux64Regex      = regexp.MustCompile("(?i)linux.amd64")

	negativeMatches = []*regexp.Regexp{appleRegex, darwinRegex, freeBSRegex, packageFileRegex, windowsRegex}
	positiveMatches = []*regexp.Regexp{architectureRegex, linux64Regex}
)

type getReleasesResp struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Url  string `json:"url"`
		Name string `json:"name"`
	} `json:"assets"`
}

type Input struct {
	base.Input
	Owner string
	Repo  string
}

func printSpec(input Input) error {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return err
	}

	owner, repo := input.Owner, input.Repo

	apiPath := fmt.Sprintf("repos/%s/%s/releases", owner, repo)
	var assetsResp []getReleasesResp
	err = client.Get(apiPath, &assetsResp)
	if err != nil {
		return err
	}

	if len(assetsResp) == 0 {
		return errors.New("no releases found")
	}

	latest := assetsResp[0]
	version := latest.TagName

	var name string
	var releaseURL string
assets:
	for _, asset := range latest.Assets {
		name = asset.Name

		for _, regex := range negativeMatches {
			if regex.MatchString(name) {
				continue assets
			}
		}

		for _, regex := range positiveMatches {
			if regex.MatchString(name) {
				releaseURL = asset.Url
				break assets
			}
		}
	}

	if releaseURL == "" {
		return errors.New("no release found")
	}
	releaseURL = fmt.Sprintf("%s/%s", version, name)
	releaseURL = strings.ReplaceAll(releaseURL, version, "${version}")

	release := entity.GithubRelease{Release: entity.Release{Ref: fmt.Sprintf("%s/%s", owner, repo), Url: releaseURL}}
	releases := []entity.GithubRelease{release}

	encoder := yaml.NewEncoder(os.Stdout)
	return encoder.Encode(&releases)
}

func PrintSpec(input Input) {
	internal.InitLogging(input.LogLevel)

	err := printSpec(input)
	if err != nil {
		internal.Logger.Fatal().Err(err).Msg("Failed to print GitHub release spec")
	}
}
