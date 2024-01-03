package entity

import (
	"github.com/femnad/fup/settings"
)

type Config struct {
	Filename        string
	Remote          bool
	AcceptHostKeys  []string          `yaml:"host_key"`
	AptRepos        []AptRepo         `yaml:"apt_repo"`
	Archives        []Release         `yaml:"archive"`
	Cargo           []CargoPkg        `yaml:"rust"`
	DnfRepos        []DnfRepo         `yaml:"dnf_repo"`
	Dirs            []DirGroup        `yaml:"dir"`
	EnsureLines     []LineInFile      `yaml:"line"`
	Flatpak         Flatpak           `yaml:"flatpak"`
	GithubUserKey   UserKey           `yaml:"github_key"`
	Go              []GoPkg           `yaml:"go"`
	Packages        PackageSpec       `yaml:"package"`
	PostflightTasks []Task            `yaml:"postflight"`
	PreflightTasks  []Task            `yaml:"preflight"`
	Python          []PythonPkg       `yaml:"python"`
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
