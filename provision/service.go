package provision

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
)

var (
	serviceDir = internal.ExpandUser("~/.config/systemd/user")
	actions    = map[string]string{
		"enable": "is-enabled",
		"start":  "is-active",
	}
	gerunds = map[string]string{
		"enable": "enabling",
		"start":  "starting",
	}
)

const svcTmpl = `[Unit]
Description={{ .Unit.Desc }}

[Service]
ExecStart={{ .Unit.Exec }}
{{- range $key, $value := .Unit.Environment  }}
Environment={{$key}}={{$value}}
{{- end }}

[Install]
WantedBy=default.target
`

type tmplOut struct {
	buf       bytes.Buffer
	sha256sum string
}

func checksum(f string) (string, error) {
	_, err := os.Stat(f)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	fd, err := os.Open(f)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	_, err = io.Copy(h, fd)
	if err != nil {
		return "", err
	}

	sum := h.Sum(nil)
	return hex.EncodeToString(sum), nil
}

func writeTmpl(s base.Service) (tmplOut, error) {
	var o tmplOut

	b := bytes.Buffer{}
	st, err := template.New("service").Parse(svcTmpl)
	if err != nil {
		return o, fmt.Errorf("error creating template: %v", err)
	}

	s.Unit.Exec = os.ExpandEnv(s.Unit.Exec)

	h := sha256.New()
	wrtSum := io.MultiWriter(&b, h)

	err = st.Execute(wrtSum, s)
	if err != nil {
		return o, fmt.Errorf("error applying service template: %v", err)
	}

	o.buf = b
	sum := h.Sum(nil)
	o.sha256sum = hex.EncodeToString(sum)

	return o, nil
}

func persist(s base.Service) error {
	if s.DontTemplate {
		return nil
	}

	var prevChecksum string

	f := fmt.Sprintf("%s/%s.service", serviceDir, s.Name)
	_, err := os.Stat(f)

	prevFile := err == nil
	if prevFile {
		prevChecksum, err = checksum(f)
		if err != nil {
			return fmt.Errorf("error checksumming existing file %s: %v", f, err)
		}
	}

	o, err := writeTmpl(s)
	if err != nil {
		return err
	}

	if prevFile && prevChecksum == o.sha256sum {
		return nil
	}

	dir, _ := path.Split(f)
	if err = common.EnsureDir(dir); err != nil {
		return err
	}

	fd, err := os.OpenFile(f, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("error opening service file %s, %v", f, err)
	}
	defer fd.Close()

	_, err = io.Copy(fd, &o.buf)
	if err != nil {
		return fmt.Errorf("error writing to file %s, %v", f, err)
	}

	internal.Log.Infof("Reloading unit files for %s", s.Name)
	_, err = common.RunCmd("systemctl --user daemon-reload")
	if err != nil {
		return fmt.Errorf("error running systemctl command: %v", err)
	}

	return nil
}

func check(s base.Service, action string) (string, error) {
	checkVerb, ok := actions[action]
	if !ok {
		return "", fmt.Errorf("unknown action: %s", action)
	}

	return fmt.Sprintf("systemctl --user %s %s", checkVerb, s.Name), nil
}

func actuate(s base.Service, action string) (string, error) {
	_, ok := actions[action]
	if !ok {
		return "", fmt.Errorf("unknown action: %s", action)
	}

	return fmt.Sprintf("systemctl --user %s %s", action, s.Name), nil
}

func ensure(s base.Service, action string) error {
	checkCmd, err := check(s, action)
	if err != nil {
		return err
	}

	_, _, r := common.RunCmdExitCode(checkCmd)
	if r == 0 {
		return nil
	}

	actuateCmd, err := actuate(s, action)
	if err != nil {
		return err
	}

	caser := cases.Title(language.Und)
	verb := caser.String(gerunds[action])
	internal.Log.Infof("%s service %s", verb, s.Name)
	_, err = common.RunCmd(actuateCmd)
	if err != nil {
		return err
	}

	return nil
}

func enable(s base.Service) error {
	if s.DontEnable {
		return nil
	}

	return ensure(s, "enable")
}

func start(s base.Service) error {
	if s.DontStart {
		return nil
	}

	return ensure(s, "start")
}

func expandService(s base.Service, cfg base.Config) (base.Service, error) {
	if s.DontTemplate {
		return s, nil
	}

	exec := s.Unit.Exec

	exec = os.Expand(exec, func(prop string) string {
		val := os.Getenv(prop)
		if val != "" {
			return val
		}

		if prop == "version" {
			return cfg.Settings.Versions[s.Name]
		}

		internal.Log.Warningf("Unable resolve property %s for service %s", prop, s.Name)
		return ""
	})

	tokens := strings.Split(exec, " ")
	if len(tokens) == 0 {
		return s, fmt.Errorf("unable to tokenize executable for service %s", s.Name)
	}
	baseExec, err := common.Which(tokens[0])
	if err != nil {
		return s, err
	}

	tokens = append([]string{baseExec}, tokens[1:]...)
	s.Unit.Exec = strings.Join(tokens, " ")

	env := s.Unit.Environment
	for k, v := range s.Unit.Environment {
		env[k] = os.ExpandEnv(v)
	}
	s.Unit.Environment = env

	return s, nil
}

func initService(s base.Service, cfg base.Config) {
	s, err := expandService(s, cfg)
	if err != nil {
		internal.Log.Errorf("error expanding service %s: %v", s.Name, err)
		return
	}

	err = persist(s)
	if err != nil {
		internal.Log.Errorf("error persisting service %s: %v", s.Name, err)
		return
	}

	err = enable(s)
	if err != nil {
		internal.Log.Errorf("error enabling service %s: %v", s.Name, err)
		return
	}

	err = start(s)
	if err != nil {
		internal.Log.Errorf("error starting service %s: %v", s.Name, err)
		return
	}
}
