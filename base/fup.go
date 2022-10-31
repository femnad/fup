package base

import (
	"gopkg.in/yaml.v3"
	"io"
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

func ReadConfig(filename string) (Config, error) {
	config := Config{}

	f, err := os.Open(filename)
	if err != nil {
		return config, err
	}

	data, err := io.ReadAll(f)
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
