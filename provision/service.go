package provision

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
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
	marecmd "github.com/femnad/mare/cmd"
)

type systemdAction struct {
	actuateCmd string
	checkCmd   string
	logVerb    string
}

var (
	systemServiceDir = "/usr/lib/systemd/system"
	userServiceDir   = internal.ExpandUser("~/.config/systemd/user")
	actions          = map[string]systemdAction{
		"disable": {actuateCmd: "disable", checkCmd: "!is-enabled", logVerb: "disabling"},
		"enable":  {actuateCmd: "enable", checkCmd: "is-enabled", logVerb: "enabling"},
		"start":   {actuateCmd: "start", checkCmd: "is-active", logVerb: "starting"},
		"stop":    {actuateCmd: "stop", checkCmd: "!is-active", logVerb: "stopping"},
	}
)

const (
	negationPrefix       = "!"
	serviceExecLineStart = "ExecStart="
	svcTmpl              = `[Unit]
Description={{ .Unit.Desc }}
{{- if .Unit.Before}}
Before={{ .Unit.Before }}.target
{{- end }}

[Service]
ExecStart={{ .Unit.Exec }}
{{- range $key, $value := .Unit.Environment }}
Environment={{$key}}={{$value}}
{{- end }}
{{- range $key, $value := .Unit.Options }}
{{ $key }}={{ $value }}
{{- end }}

[Install]
WantedBy={{ if .Unit.WantedBy }}{{ .Unit.WantedBy }}{{ else }}{{ "default" }}{{ end }}.target
`
)

func writeTmpl(s base.Service) (string, error) {
	b := bytes.Buffer{}

	options := make(map[string]string)
	for k, v := range s.Unit.Options {
		options[k] = os.ExpandEnv(v)
	}
	s.Unit.Options = options

	st, err := template.New("service").Parse(svcTmpl)
	if err != nil {
		return "", fmt.Errorf("error creating template: %v", err)
	}

	err = st.Execute(&b, s)
	if err != nil {
		return "", fmt.Errorf("error applying service template: %v", err)
	}

	return b.String(), nil
}

func runSystemctlCmd(cmd string, service base.Service) error {
	internal.Log.Debugf("running systemctl command %s for service %s", cmd, service.Name)
	_, err := marecmd.RunFormatError(marecmd.Input{Command: cmd, Sudo: service.System})
	return err
}

func getServiceFilePath(s base.Service) string {
	if s.System {
		return fmt.Sprintf("%s/%s.service", systemServiceDir, s.Name)
	}

	return fmt.Sprintf("%s/%s.service", userServiceDir, s.Name)
}

func writeServiceFile(file, content string) (bool, error) {
	return internal.WriteContent(internal.ManagedFile{Path: file, Content: content})
}

func maybeRestart(s base.Service) error {
	cmd := systemctlCmd("is-active", s.Name, !s.System)
	out, err := marecmd.RunFormatError(marecmd.Input{Command: cmd})
	if err != nil {
		if strings.TrimSpace(out.Stdout) == "inactive" {
			return nil
		}
		return err
	}

	internal.Log.Debugf("restarting active service %s due to service file content changes", s.Name)

	cmd = systemctlCmd("restart", s.Name, !s.System)
	return marecmd.RunNoOutput(marecmd.Input{Command: cmd, Sudo: s.System})
}

func getServiceExec(serviceFile string) (string, error) {
	f, err := os.Open(serviceFile)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, serviceExecLineStart) {
			continue
		}

		fields := strings.Split(line, "=")
		return fields[1], nil
	}

	return "", nil
}

func persist(s base.Service) error {
	if s.DontTemplate {
		return nil
	}

	serviceFilePath := getServiceFilePath(s)
	if !s.System {
		dir, _ := path.Split(serviceFilePath)
		if err := internal.EnsureDir(dir); err != nil {
			return err
		}
	}

	prevExec, err := getServiceExec(serviceFilePath)
	if err != nil {
		return err
	}

	o, err := writeTmpl(s)
	if err != nil {
		return err
	}

	changed, err := writeServiceFile(serviceFilePath, o)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}

	if changed && !internal.IsHomePath(serviceFilePath) {
		isRoot, err := internal.IsUserRoot()
		if err != nil {
			return err
		}

		// Fix "SELinux is preventing systemd from open access on the file <service-file>" error
		restorecon := fmt.Sprintf("/sbin/restorecon %s", serviceFilePath)
		internal.Log.Debugf("running restorecon command %s", restorecon)

		_, cmdErr := marecmd.RunFormatError(marecmd.Input{Command: restorecon, Sudo: !isRoot})
		if cmdErr != nil {
			return cmdErr
		}
	}

	internal.Log.Infof("Reloading unit files for %s", s.Name)
	c := systemctlCmd("daemon-reload", "", !s.System)
	err = runSystemctlCmd(c, s)
	if err != nil {
		return err
	}

	if prevExec == "" || prevExec == s.Unit.Exec {
		return nil
	}

	return maybeRestart(s)
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

func check(s base.Service, action systemdAction) (string, bool, error) {
	var negated bool

	checkAction := action.checkCmd
	negated = strings.HasPrefix(checkAction, negationPrefix)
	if negated {
		checkAction = strings.TrimLeft(checkAction, negationPrefix)
	}

	return systemctlCmd(checkAction, s.Name, !s.System), negated, nil
}

func actuate(s base.Service, action systemdAction) (string, error) {
	return systemctlCmd(action.actuateCmd, s.Name, !s.System), nil
}

func ensure(s base.Service, actionStr string) error {
	action, ok := actions[actionStr]
	if !ok {
		return fmt.Errorf("no such action: %s", actionStr)
	}

	checkCmd, negated, err := check(s, action)
	if err != nil {
		return err
	}

	// Don't need sudo for check actions, so don't use runSystemctlCmd
	resp, _ := marecmd.RunFormatError(marecmd.Input{Command: checkCmd})
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
	verb := caser.String(action.logVerb)
	internal.Log.Infof("%s service %s", verb, s.Name)

	return runSystemctlCmd(actuateCmd, s)
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
	exec = settings.ExpandStringWithLookup(cfg.Settings, exec, lookup)

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

func initService(s base.Service, cfg base.Config) error {
	if !when.ShouldRun(s) {
		internal.Log.Debugf("Skipping initializing %s as when condition %s evaluated to false", s.Name, s.When)
		return nil
	}

	if s.Stop {
		err := ensure(s, "stop")
		if err != nil {
			internal.Log.Errorf("error stopping service %s, %v", s.Name, err)
		}
		return err
	}

	if s.Disable {
		err := ensure(s, "disable")
		if err != nil {
			internal.Log.Errorf("error disabling service %s, %v", s.Name, err)
		}
		return err
	}

	s, err := expandService(s, cfg)
	if err != nil {
		internal.Log.Errorf("error expanding service %s: %v", s.Name, err)
		return err
	}

	err = persist(s)
	if err != nil {
		internal.Log.Errorf("error persisting service %s: %v", s.Name, err)
		return err
	}

	err = enable(s)
	if err != nil {
		internal.Log.Errorf("error enabling service %s: %v", s.Name, err)
		return err
	}

	err = start(s)
	if err != nil {
		internal.Log.Errorf("error starting service %s: %v", s.Name, err)
		return err
	}

	return nil
}

func initServices(config base.Config) error {
	var svcErr []error
	for _, svc := range config.Services {
		err := initService(svc, config)
		svcErr = append(svcErr, err)
	}

	return errors.Join(svcErr...)
}
