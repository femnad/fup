package base

import (
	"fmt"
	"github.com/femnad/fup/remote"
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"os"
)

type Settings struct {
	ArchiveDir string                    `yaml:"archive_dir"`
	HostFacts  map[string]map[string]any `yaml:"host_facts"`
}

type Unless struct {
	Cmd  string `yaml:"cmd"`
	Post string `yaml:"post"`
	Ls   string `yaml:"ls"`
}

func (u Unless) String() string {
	if u.Ls != "" {
		return fmt.Sprintf("ls %s", u.Ls)
	}

	s := u.Cmd
	if u.Post != "" {
		s += " | " + u.Post
	}
	return s
}

type Archive struct {
	Url     string   `yaml:"url"`
	Unless  Unless   `yaml:"unless"`
	Version string   `yaml:"version"`
	Symlink []string `yaml:"symlink"`
	Binary  string   `yaml:"binary"`
}

func (a Archive) ExpandArchive(property string) string {
	if property == "version" {
		return a.Version
	}

	return ""
}

type Config struct {
	Settings *Settings `yaml:"settings"`
	Archives []Archive `yaml:"archives"`
}

func readLocalConfigFile(config string) (io.Reader, error) {
	f, err := os.Open(config)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func readRemoteConfigFile(config string) (io.Reader, error) {
	body, err := remote.ReadResponseBody(config)
	if err != nil {
		return nil, err
	}

	return body, nil
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
	return decodeConfig(filename)
}
