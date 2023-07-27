package base

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/remote"
)

type PackageSpec map[string][]string
type RemotePackageSpec map[string][]entity.RemotePackage

type Config struct {
	AcceptHostKeys   []string            `yaml:"accept_host_keys"`
	Archives         []Archive           `yaml:"archives"`
	Binaries         []entity.Binary     `yaml:"binaries"`
	Cargo            []CargoPkg          `yaml:"rust"`
	EnsureDirs       []string            `yaml:"ensure_dirs"`
	EnsureLines      []LineInFile        `yaml:"ensure_lines"`
	GithubUserKey    UserKey             `yaml:"github_user_keys"`
	Go               []GoPkg             `yaml:"go"`
	Packages         PackageSpec         `yaml:"packages"`
	PostflightTasks  []Task              `yaml:"postflight"`
	PreflightTasks   []Task              `yaml:"preflight"`
	Python           []PythonPkg         `yaml:"python"`
	RemotePackages   RemotePackageSpec   `yaml:"remote_packages"`
	SelfRepos        []entity.Repo       `yaml:"self_repos"`
	Services         []Service           `yaml:"services"`
	Settings         settings.Settings   `yaml:"settings"`
	Tasks            []Task              `yaml:"tasks"`
	Templates        []Template          `yaml:"template"`
	UserInGroup      map[string][]string `yaml:"user_in_group"`
	UnwantedDirs     []string            `yaml:"unwanted_dirs"`
	UnwantedPackages PackageSpec         `yaml:"unwanted_packages"`
}

func readLocalConfigFile(config string) (io.Reader, error) {
	f, err := os.Open(config)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func readRemoteConfigFile(config string) (io.Reader, error) {
	response, err := remote.ReadResponseBody(config)
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

func getConfigReader(config string) (io.Reader, error) {
	parsed, err := url.Parse(config)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme == "" {
		return readLocalConfigFile(config)
	}

	return readRemoteConfigFile(config)
}

func decodeConfig(filename string) (Config, error) {
	config := Config{}
	reader, err := getConfigReader(filename)
	if err != nil {
		return config, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("error deserializing config from %s: %v", filename, err)
	}

	return config, nil
}

func ReadConfig(filename string) (Config, error) {
	filename = internal.ExpandUser(filename)
	return decodeConfig(filename)
}
