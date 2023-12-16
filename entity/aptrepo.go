package entity

import (
	"fmt"
	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/remote"
	marecmd "github.com/femnad/mare/cmd"
	"io"
	"os/exec"
	"path"
	"strings"
)

const (
	keyRingsDir = "/etc/apt/keyrings"
	sourcesDir  = "/etc/apt/sources.list.d"
)

type AptRepo struct {
	GPGKey   string `yaml:"gpg_key"`
	RepoName string `yaml:"name"`
	Repo     string `yaml:"repo"`
	When     string `yaml:"when"`
}

func (a AptRepo) DefaultVersionCmd() string {
	return ""
}

func (a AptRepo) GetUnless() unless.Unless {
	return unless.Unless{
		Stat: path.Join(sourcesDir, fmt.Sprintf("%s.list", a.RepoName)),
	}
}

func (AptRepo) GetVersion() string {
	return ""
}

func (AptRepo) HasPostProc() bool {
	return false
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

	_, err = io.ReadAll(stdout)
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
	err := internal.EnsureDir(keyRingsDir)
	if err != nil {
		return err
	}
	keyRingFile := path.Join(keyRingsDir, fmt.Sprintf("%s.gpg", a.RepoName))

	err = a.ensureKeyFile(a.GPGKey, keyRingFile)
	if err != nil {
		return err
	}

	out, err := marecmd.RunFormatError(marecmd.Input{Command: "dpkg --print-architecture"})
	if err != nil {
		return err
	}
	architecture := strings.TrimSpace(out.Stdout)

	versionCodename, err := precheck.GetOSVersionCodename()
	if err != nil {
		return err
	}

	content := fmt.Sprintf("deb [arch=${architecture} signed-by=%s] %s ${codename} stable", keyRingFile, a.Repo)
	content = settings.Expand(content, map[string]string{
		"architecture": architecture,
		"codename":     versionCodename,
	})
	repoFile := path.Join(sourcesDir, fmt.Sprintf("%s.list", a.RepoName))

	_, err = internal.WriteContent(internal.ManagedFile{
		Content: content,
		Path:    repoFile,
	})

	return err
}
