package entity

type GithubRelease struct {
	ExecName string `yaml:"exec-name,omitempty"`
	Release  `yaml:",inline"`
}
