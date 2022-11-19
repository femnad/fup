package base

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/remote"
)

type Settings struct {
	ExtractDir string                    `yaml:"extract_dir"`
	HostFacts  map[string]map[string]any `yaml:"host_facts"`
}

type PackageSpec map[string][]string

type Unless struct {
	Cmd  string `yaml:"cmd"`
	Post string `yaml:"post"`
	Stat string `yaml:"stat"`
}

func (u Unless) HasPostProc() bool {
	return u.Post == "" || u.Cmd == ""
}

func (u Unless) String() string {
	if u.Stat != "" {
		return fmt.Sprintf("ls %s", u.Stat)
	}

	s := u.Cmd
	if u.Post != "" {
		s += " | " + u.Post
	}
	return s
}

type Config struct {
	Archives       []Archive   `yaml:"archives"`
	Packages       PackageSpec `yaml:"packages"`
	PreflightTasks []Task      `yaml:"preflight"`
	Settings       Settings    `yaml:"settings"`
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
		return config, err
	}

	return config, nil
}

func ReadConfig(filename string) (Config, error) {
	filename = internal.ExpandUser(filename)
	return decodeConfig(filename)
}
