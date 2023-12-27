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
	file            string
	isRemote        bool
	AcceptHostKeys  []string           `yaml:"host_key"`
	AptRepos        []entity.AptRepo   `yaml:"apt_repo"`
	Archives        []Archive          `yaml:"archive"`
	Binaries        []entity.Binary    `yaml:"binary"`
	Cargo           []CargoPkg         `yaml:"rust"`
	DnfRepos        []entity.DnfRepo   `yaml:"dnf_repo"`
	EnsureDirs      []entity.Dir       `yaml:"dir"`
	EnsureLines     []LineInFile       `yaml:"line"`
	Flatpak         entity.Flatpak     `yaml:"flatpak"`
	GithubUserKey   UserKey            `yaml:"github_key"`
	Go              []GoPkg            `yaml:"go"`
	Packages        PackageSpec        `yaml:"package"`
	PostflightTasks []Task             `yaml:"postflight"`
	PreflightTasks  []Task             `yaml:"preflight"`
	Python          []PythonPkg        `yaml:"python"`
	RemotePackages  RemotePackageSpec  `yaml:"remote_package"`
	RepoGroups      []entity.RepoGroup `yaml:"repo"`
	Services        []Service          `yaml:"service"`
	Settings        settings.Settings  `yaml:"settings"`
	SnapPackages    []entity.Snap      `yaml:"snap"`
	Tasks           []Task             `yaml:"task"`
	Templates       []Template         `yaml:"template"`
	UserInGroup     UserInGroupSpec    `yaml:"user_group"`
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
