package entity

type GithubRelease struct {
	ExecName string `yaml:"exec-name"`
	Release  `yaml:",inline"`
}
