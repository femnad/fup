package entity

import (
	"fmt"
	"io"
	"os/exec"
	"path"
	"strings"

	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/remote"
	"github.com/femnad/fup/settings"
	marecmd "github.com/femnad/mare/cmd"
)

const (
	defaultComponents = "stable"
	keyRingsDir       = "/etc/apt/keyrings"
	sourcesDir        = "/etc/apt/sources.list.d"
)

type AptRepo struct {
	unless.BasicUnlessable
	Distribution string `yaml:"distribution"`
	Components   string `yaml:"components"`
	GPGKey       string `yaml:"gpg_key"`
	Pin          bool   `yaml:"pin"`
	RepoName     string `yaml:"name"`
	Repo         string `yaml:"repo"`
	When         string `yaml:"when"`
}

func (a AptRepo) DefaultVersionCmd() string {
	return ""
}

func (a AptRepo) GetUnless() unless.Unless {
	return unless.Unless{
		Stat: path.Join(sourcesDir, fmt.Sprintf("%s.list", a.RepoName)),
	}
}

func (AptRepo) LookupVersion(_ settings.Settings) (string, error) {
	return "", nil
}

func (a AptRepo) Name() string {
	return a.RepoName
}

func (a AptRepo) RunWhen() string {
	return a.When
}

func (AptRepo) UpdateCmd() string {
	return "apt update"
}

func (AptRepo) ensureKeyFile(keyUrl, keyRingFile string) error {
	key, err := remote.ReadResponseBytes(keyUrl)
	if err != nil {
		return err
	}

	gpgCmd := exec.Command("gpg", "--dearmor")

	stdin, err := gpgCmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	stdout, err := gpgCmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()

	stderr, err := gpgCmd.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()

	if err = gpgCmd.Start(); err != nil {
		return err
	}

	_, err = stdin.Write(key)
	if err != nil {
		return err
	}
	err = stdin.Close()
	if err != nil {
		return err
	}

	gpgKey, err := io.ReadAll(stdout)
	if err != nil {
		return err
	}

	if err = gpgCmd.Wait(); err != nil {
		out, err := io.ReadAll(stderr)
		if err != nil {
			return err
		}

		return fmt.Errorf("error in gpg process, output %s: %v", out, err)
	}

	_, err = internal.WriteContent(internal.ManagedFile{
		Content: string(gpgKey),
		Path:    keyRingFile,
		Mode:    0o644,
		User:    "root",
		Group:   "root",
	})
	return err
}

func (a AptRepo) Install() error {
	err := internal.EnsureDirExists(keyRingsDir)
	if err != nil {
		return err
	}
	keyRingFile := path.Join(keyRingsDir, fmt.Sprintf("%s.gpg", a.RepoName))

	err = a.ensureKeyFile(a.GPGKey, keyRingFile)
	if err != nil {
		return err
	}

	out, err := marecmd.RunFmtErr(marecmd.Input{Command: "dpkg --print-architecture"})
	if err != nil {
		return err
	}
	architecture := strings.TrimSpace(out.Stdout)

	distribution := a.Distribution
	if distribution == "" {
		distribution, err = precheck.GetOSVersionCodename()
		if err != nil {
			return err
		}
	}

	components := a.Components
	if components == "" {
		components = defaultComponents
	}

	name := a.RepoName
	repo := a.Repo
	content := fmt.Sprintf("deb [arch=%s signed-by=%s] %s %s %s\n", architecture, keyRingFile,
		repo, distribution, components)
	repoFile := path.Join(sourcesDir, fmt.Sprintf("%s.list", name))

	_, err = internal.WriteContent(internal.ManagedFile{
		Content: content,
		Path:    repoFile,
	})
	if err != nil {
		return err
	}

	if !a.Pin {
		return nil
	}

	origin, _ := path.Split(repo)
	fields := strings.Split(origin, "://")
	if len(fields) < 2 {
		return fmt.Errorf("unable to determine origin from repository %s", repo)
	}

	origin = strings.TrimSuffix(fields[1], "/")
	pinContent := fmt.Sprintf(`Package: *
Pin: origin %s
Pin-Priority: 1000`, origin)
	pinFile := fmt.Sprintf("/etc/apt/preferences.d/%s", name)
	_, err = internal.WriteContent(internal.ManagedFile{
		Content: pinContent,
		Path:    pinFile,
	})
	return err
}
