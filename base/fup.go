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

type PackageSpec []entity.PackageGroup
type RemotePackageSpec []entity.RemotePackageGroup
type UserInGroupSpec map[string][]entity.Group

type Config struct {
	file             string
	isRemote         bool
	AcceptHostKeys   []string          `yaml:"accept_host_keys"`
	Archives         []Archive         `yaml:"archives"`
	Binaries         []entity.Binary   `yaml:"binaries"`
	Cargo            []CargoPkg        `yaml:"rust"`
	EnsureDirs       []string          `yaml:"ensure_dirs"`
	EnsureLines      []LineInFile      `yaml:"ensure_lines"`
	Flatpak          entity.Flatpak    `yaml:"flatpak"`
	GithubUserKey    UserKey           `yaml:"github_user_keys"`
	Go               []GoPkg           `yaml:"go"`
	Packages         PackageSpec       `yaml:"packages"`
	PostflightTasks  []Task            `yaml:"postflight"`
	PreflightTasks   []Task            `yaml:"preflight"`
	Python           []PythonPkg       `yaml:"python"`
	RemotePackages   RemotePackageSpec `yaml:"remote_packages"`
	Repos            []entity.Repo     `yaml:"repos"`
	Services         []Service         `yaml:"services"`
	Settings         settings.Settings `yaml:"settings"`
	SnapPackages     []entity.Snap     `yaml:"snap"`
	Tasks            []Task            `yaml:"tasks"`
	Templates        []Template        `yaml:"template"`
	UserInGroup      UserInGroupSpec   `yaml:"user_in_group"`
	UnwantedDirs     []string          `yaml:"unwanted_dirs"`
	UnwantedPackages PackageSpec       `yaml:"unwanted_packages"`
}

func (c Config) IsRemote() bool {
	return c.isRemote
}

func (c Config) File() string {
	return c.file
}

type configReader struct {
	reader   io.Reader
	isRemote bool
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

func getConfigReader(config string) (configReader, error) {
	parsed, err := url.Parse(config)
	if err != nil {
		return configReader{}, err
	}

	var readerFn func(string) (io.Reader, error)
	var isRemote bool
	if parsed.Scheme == "" {
		readerFn = readLocalConfigFile
		isRemote = false
	} else {
		readerFn = readRemoteConfigFile
		isRemote = true
	}

	reader, err := readerFn(config)
	if err != nil {
		return configReader{}, err
	}

	return configReader{reader: reader, isRemote: isRemote}, nil
}

func decodeConfig(filename string) (Config, error) {
	config := Config{}
	cfgReader, err := getConfigReader(filename)
	if err != nil {
		return config, err
	}

	data, err := io.ReadAll(cfgReader.reader)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("error deserializing config from %s: %v", filename, err)
	}

	config.isRemote = cfgReader.isRemote
	config.file = filename
	return config, nil
}

func ReadConfig(filename string) (Config, error) {
	filename = internal.ExpandUser(filename)
	return decodeConfig(filename)
}
