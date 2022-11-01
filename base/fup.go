package base

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"

	"gopkg.in/yaml.v3"

	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/remote"
)

type Settings struct {
	ExtractDir string                    `yaml:"extract_dir"`
	HostFacts  map[string]map[string]any `yaml:"host_facts"`
}

type Unless struct {
	Cmd  string `yaml:"cmd"`
	Post string `yaml:"post"`
	Stat string `yaml:"stat"`
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

type Archive struct {
	Url     string   `yaml:"url"`
	Unless  Unless   `yaml:"unless"`
	Version string   `yaml:"version"`
	Symlink []string `yaml:"symlink"`
	Binary  string   `yaml:"binary"`
}

func (a Archive) String() string {
	return a.Url
}

func (a Archive) expand(property string) string {
	if property == "version" {
		return a.Version
	}

	return ""
}

func (a Archive) ExpandURL() string {
	return os.Expand(a.Url, a.expand)
}

func (a Archive) ShortURL() string {
	_, basename := path.Split(a.ExpandURL())
	return basename
}

func (a Archive) ExpandSymlinks() []string {
	var expanded []string
	for _, symlink := range a.Symlink {
		expanded = append(expanded, a.expand(symlink))
	}

	return expanded
}

type Config struct {
	Settings Settings  `yaml:"settings"`
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
	filename = internal.ExpandUser(filename)
	return decodeConfig(filename)
}
