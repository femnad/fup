package base

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/remote"
)

type PackageSpec map[string][]string

type Config struct {
	AcceptHostKeys   []string          `yaml:"accept_host_keys"`
	Archives         []Archive         `yaml:"archives"`
	Cargo            []CargoPkg        `yaml:"cargo"`
	EnsureDirs       []string          `yaml:"ensure_dirs"`
	GithubUserKey    UserKey           `yaml:"github_user_keys"`
	Go               []GoPkg           `yaml:"go"`
	Packages         PackageSpec       `yaml:"packages"`
	PostflightTasks  []Task            `yaml:"postflight"`
	PreflightTasks   []Task            `yaml:"preflight"`
	Python           []PythonPkg       `yaml:"python"`
	Services         []Service         `yaml:"services"`
	Settings         settings.Settings `yaml:"settings"`
	Tasks            []Task            `yaml:"tasks"`
	Templates        []Template        `yaml:"template"`
	UnwantedDirs     []string          `yaml:"unwanted_dirs"`
	UnwantedPackages PackageSpec       `yaml:"unwanted_packages"`
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
