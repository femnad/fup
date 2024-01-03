package entity

import (
	"github.com/femnad/fup/settings"
)

type Config struct {
	Filename        string
	Remote          bool
	AcceptHostKeys  []string          `yaml:"host_key"`
	AptRepos        []AptRepo         `yaml:"apt_repo"`
	Cargo           []CargoPkg        `yaml:"rust"`
	Dirs            []DirGroup        `yaml:"dir"`
	DnfRepos        []DnfRepo         `yaml:"dnf_repo"`
	EnsureLines     []LineInFile      `yaml:"line"`
	Flatpak         Flatpak           `yaml:"flatpak"`
	GithubUserKey   UserKey           `yaml:"github_key"`
	Go              []GoPkg           `yaml:"go"`
	Packages        PackageSpec       `yaml:"package"`
	PostflightTasks []Task            `yaml:"postflight"`
	PreflightTasks  []Task            `yaml:"preflight"`
	Python          []PythonPkg       `yaml:"python"`
	Releases        []Release         `yaml:"release"`
	RemotePackages  RemotePackageSpec `yaml:"remote_package"`
	Repos           []Repo            `yaml:"repo"`
	Services        []Service         `yaml:"service"`
	Settings        settings.Settings `yaml:"settings"`
	SnapPackages    []Snap            `yaml:"snap"`
	Tasks           []Task            `yaml:"task"`
	Templates       []Template        `yaml:"template"`
	UserInGroup     UserInGroupSpec   `yaml:"user_group"`
}

type PackageSpec []PackageGroup
type RemotePackageSpec []RemotePackageGroup
type UserInGroupSpec map[string][]Group

func (c Config) IsRemote() bool {
	return c.Remote
}

func (c Config) File() string {
	return c.Filename
}
