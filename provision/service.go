package provision

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

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

const tmpl = `[Unit]
Description={{ .Unit.Desc }}

[Service]
ExecStart={{ .Unit.Exec }}
{{- range $key, $value := .Unit.Environment  }}
Environment={{$key}}={{$value}}
{{- end }}

[Install]
WantedBy=default.target
`

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

func write(s base.Service, f string) error {
	fd, err := os.OpenFile(f, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("error opening service file %s, %v", f, err)
	}
	defer fd.Close()

	st, err := template.New("service").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("error creating template: %v", err)
	}

	s.Unit.Exec = os.ExpandEnv(s.Unit.Exec)

	err = st.Execute(fd, s)
	if err != nil {
		return fmt.Errorf("error applying service template: %v", err)
	}

	return nil
}

func persist(s base.Service) error {
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

	err = write(s, f)
	if err != nil {
		return err
	}

	if !prevFile {
		return nil
	}

	sum, err := checksum(f)
	if err != nil {
		return fmt.Errorf("error checksumming new file %s: %v", f, err)
	}

	if prevChecksum == sum {
		return nil
	}

	internal.Log.Infof("Reloading unit files for %s", s.Name)
	common.RunCmd("systemctl --user daemon-reload")
	return nil
}

func check(s base.Service, action string) (string, error) {
	check, ok := actions[action]
	if !ok {
		return "", fmt.Errorf("unknown action: %s", action)
	}

	return fmt.Sprintf("systemctl --user %s %s", check, s.Name), nil
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

	r := common.RunCmdExitCode(checkCmd)
	if r == 0 {
		return nil
	}

	actuateCmd, err := actuate(s, action)
	if err != nil {
		return err
	}

	verb := strings.Title(gerunds[action])
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

	tokens = append([]string{baseExec}, tokens[1:len(tokens)]...)
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
