package entity

type GithubRelease struct {
	Release  `yaml:",inline"`
	ExecName string `yaml:"exec-name"`
}
