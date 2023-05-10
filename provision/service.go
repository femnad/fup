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
	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

var (
	systemServiceDir = internal.ExpandUser("/usr/lib/systemd/system")
	userServiceDir   = internal.ExpandUser("~/.config/systemd/user")
	actions          = map[string]string{
		"disable": "!is-enabled",
		"enable":  "is-enabled",
		"start":   "is-active",
	}
)

const (
	negationPrefix = "!"
	svcTmpl        = `[Unit]
Description={{ .Unit.Desc }}

[Service]
ExecStart={{ .Unit.Exec }}
{{- range $key, $value := .Unit.Environment  }}
Environment={{$key}}={{$value}}
{{- end }}

[Install]
WantedBy=default.target
`
)

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

func runSystemctlCmd(cmd string, service base.Service) (common.CmdOut, error) {
	internal.Log.Debugf("running systemctl command %s for service %s", cmd, service.Name)
	return common.RunCmd(common.CmdIn{Command: cmd, Sudo: service.System})
}

func getServiceFilePath(s base.Service) string {
	if s.System {
		return fmt.Sprintf("%s/%s.service", systemServiceDir, s.Name)
	}

	return fmt.Sprintf("%s/%s.service", userServiceDir, s.Name)
}

func persist(s base.Service) error {
	if s.DontTemplate {
		return nil
	}

	var prevChecksum string

	serviceFilePath := getServiceFilePath(s)
	_, err := os.Stat(serviceFilePath)

	prevFile := err == nil
	if prevFile {
		prevChecksum, err = checksum(serviceFilePath)
		if err != nil {
			return fmt.Errorf("error checksumming existing file %s: %v", serviceFilePath, err)
		}
	}

	o, err := writeTmpl(s)
	if err != nil {
		return err
	}

	if prevFile && prevChecksum == o.sha256sum {
		return nil
	}

	dir, _ := path.Split(serviceFilePath)
	if err = common.EnsureDir(dir); err != nil {
		return err
	}

	fd, err := os.OpenFile(serviceFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("error opening service file %s, %v", serviceFilePath, err)
	}
	defer fd.Close()

	_, err = io.Copy(fd, &o.buf)
	if err != nil {
		return fmt.Errorf("error writing to file %s, %v", serviceFilePath, err)
	}

	internal.Log.Infof("Reloading unit files for %s", s.Name)
	c := systemctlCmd("daemon-reload", "", !s.System)
	_, err = runSystemctlCmd(c, s)
	if err != nil {
		return fmt.Errorf("error running systemctl command: %v", err)
	}

	return nil
}

func systemctlCmd(action, target string, user bool) string {
	var maybeUser string
	var maybeTarget string

	if user {
		maybeUser = "--user "
	}
	if target != "" {
		maybeTarget = " " + target
	}

	return fmt.Sprintf("systemctl %s%s%s", maybeUser, action, maybeTarget)
}

func check(s base.Service, action string) (string, bool, error) {
	var negated bool

	checkVerb, ok := actions[action]
	if !ok {
		return "", negated, fmt.Errorf("unknown action: %s", action)
	}

	negated = strings.HasPrefix(checkVerb, negationPrefix)
	if negated {
		checkVerb = strings.TrimLeft(checkVerb, negationPrefix)
	}

	return systemctlCmd(checkVerb, s.Name, !s.System), negated, nil
}

func actuate(s base.Service, action string) (string, error) {
	_, ok := actions[action]
	if !ok {
		return "", fmt.Errorf("unknown action: %s", action)
	}

	return systemctlCmd(action, s.Name, !s.System), nil
}

func ensure(s base.Service, action string) error {
	checkCmd, negated, err := check(s, action)
	if err != nil {
		return err
	}

	// Don't need sudo for check actions, so don't use runSystemctlCmd
	resp, _ := common.RunCmd(common.CmdIn{Command: checkCmd})
	if negated && resp.Code != 0 {
		return nil
	} else if !negated && resp.Code == 0 {
		return nil
	}

	actuateCmd, err := actuate(s, action)
	if err != nil {
		return err
	}

	caser := cases.Title(language.Und)
	verb := caser.String(action)
	internal.Log.Infof("%s-ing service %s", verb, s.Name)
	resp, err = runSystemctlCmd(actuateCmd, s)
	if err != nil {
		return fmt.Errorf("error %s-ing service %s: output: %s error: %v", action, s.Name, resp.Stderr, err)
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

	lookup := map[string]string{"version": cfg.Settings.Versions[s.Name]}
	exec = settings.ExpandString(cfg.Settings, exec, lookup)

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
	if !when.ShouldRun(s) {
		internal.Log.Debugf("Skipping initializing %s as when condition %s evaluated to false", s.Name, s.When)
		return
	}

	if s.Disable {
		err := ensure(s, "disable")
		if err != nil {
			internal.Log.Errorf("error disabling service %s, %v", s.Name, err)
		}
		return
	}

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
