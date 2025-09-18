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

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
	"github.com/femnad/fup/settings"
	marecmd "github.com/femnad/mare/cmd"
)

type systemdAction struct {
	actuateCmd string
	checkCmd   string
	logVerb    string
}

var (
	restoreConExec   = "/sbin/restorecon"
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
	oneshotService       = "oneshot"
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
	timerTmpl = `[Unit]
Description={{ .Timer.Desc }}

[Timer]
OnCalendar={{ .Timer.Calendar }}
RandomizedDelaySec={{ .Timer.RandomizedDelay }}
Persistent=true

[Install]
WantedBy=timers.target`
)

func writeTmpl(s entity.Service) (string, error) {
	b := bytes.Buffer{}

	options := make(map[string]string)
	for k, v := range s.Unit.Options {
		options[k] = os.ExpandEnv(v)
	}
	if s.Unit.Type != "" {
		options["Type"] = s.Unit.Type
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

func writeTimerTmpl(s entity.Service) (string, error) {
	buf := bytes.Buffer{}

	tt, err := template.New("timer").Parse(timerTmpl)
	if err != nil {
		return "", fmt.Errorf("error creating template: %v", err)
	}

	err = tt.Execute(&buf, s)
	if err != nil {
		return "", fmt.Errorf("error applying timer template: %v", err)
	}

	return buf.String(), nil
}

func runSystemctlCmd(cmd string, service entity.Service) error {
	internal.Logger.Trace().Str("cmd", cmd).Str("service", service.Name).Msg("Running systemctl")
	err := marecmd.RunErrOnly(marecmd.Input{Command: cmd, Sudo: service.System})
	return err
}

func getUnitFilePath(s entity.Service, unitType string) string {
	if s.System {
		return fmt.Sprintf("%s/%s.%s", systemServiceDir, s.Name, unitType)
	}

	return fmt.Sprintf("%s/%s.%s", userServiceDir, s.Name, unitType)
}

func getServiceFilePath(s entity.Service) string {
	return getUnitFilePath(s, "service")
}

func getTimerFilePath(s entity.Service) string {
	return getUnitFilePath(s, "timer")
}

func writeUnitFile(file, content string) (bool, error) {
	return internal.WriteContent(internal.ManagedFile{Path: file, Content: content})
}

func maybeRestart(s entity.Service, unitType string) error {
	if s.Unit.Type == oneshotService {
		return nil
	}

	cmd := systemctlCmd("is-active", s.Name, unitType, !s.System)
	out, err := marecmd.RunFmtErr(marecmd.Input{Command: cmd})
	if err != nil {
		if strings.TrimSpace(out.Stdout) == "inactive" {
			return nil
		}
		return err
	}

	internal.Logger.Trace().Str("name", s.Name).Msg("Restarting service due to file content changes")
	cmd = systemctlCmd("restart", s.Name, unitType, !s.System)
	return marecmd.RunErrOnly(marecmd.Input{Command: cmd, Sudo: s.System})
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

func maybeRunRestoreCon(serviceFilePath string) error {
	if _, err := os.Stat(restoreConExec); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	// Fix "SELinux is preventing systemd from open access on the file <service-file>" error
	cmd := fmt.Sprintf("%s %s", restoreConExec, serviceFilePath)
	return internal.MaybeRunWithSudo(cmd)
}

func persistUnit(s entity.Service) (restart bool, err error) {
	if s.Unit == nil {
		return false, nil
	}

	name := s.Name
	execFields := strings.Split(s.Unit.Exec, " ")
	if len(execFields) == 0 {
		return restart, fmt.Errorf("unable to determine executable for service %s", name)
	}

	exec := execFields[0]
	info, err := os.Stat(exec)
	if err != nil {
		return restart, fmt.Errorf("error looking up executable for service %s: %v", name, err)
	}

	if !common.IsExecutableFile(info) {
		return restart, fmt.Errorf("executable %s for service %s does not point to an executable file", exec, name)
	}

	if s.Unit.Desc == "" {
		return restart, fmt.Errorf("description required for templating service %s", name)
	}

	serviceFilePath := getServiceFilePath(s)
	if !s.System {
		dir, _ := path.Split(serviceFilePath)
		if err = internal.EnsureDirExists(dir); err != nil {
			return
		}
	}

	prevExec, err := getServiceExec(serviceFilePath)
	if err != nil {
		return
	}
	newExec := prevExec != "" && prevExec != s.Unit.Exec

	o, err := writeTmpl(s)
	if err != nil {
		return
	}

	changed, err := writeUnitFile(serviceFilePath, o)
	if err != nil {
		return
	}

	restart = changed || newExec

	if restart && !internal.IsHomePath(serviceFilePath) {
		err = maybeRunRestoreCon(serviceFilePath)
		if err != nil {
			return restart, err
		}
	}

	return
}

func maybePersistTimer(s entity.Service) (bool, error) {
	if s.Timer == nil {
		return false, nil
	}

	o, err := writeTimerTmpl(s)
	if err != nil {
		return false, err
	}

	timerFilePath := getTimerFilePath(s)
	return writeUnitFile(timerFilePath, o)
}

func reload(s entity.Service, unitType string) error {
	internal.Logger.Trace().Str("name", s.Name).Msg("Reloading unit files")
	c := systemctlCmd("daemon-reload", "", unitType, !s.System)
	return runSystemctlCmd(c, s)
}

func persist(s entity.Service) error {
	if s.DontTemplate {
		return nil
	}

	restartService, err := persistUnit(s)
	if err != nil {
		return err
	}

	restartTimer, err := maybePersistTimer(s)
	if err != nil {
		return err
	}

	if s.Timer == nil {
		if !restartService {
			return nil
		}
	} else if !restartService && !restartTimer {
		return nil
	}

	err = reload(s, "service")
	if err != nil {
		return err
	}
	if s.Timer != nil {
		err = reload(s, "timer")
	}

	err = maybeRestart(s, "service")
	if err != nil {
		return err
	}
	if s.Timer != nil {
		err = maybeRestart(s, "timer")
		if err != nil {
			return err
		}
	}

	return nil
}

func systemctlCmd(action, target, unitType string, user bool) string {
	var maybeUser string
	var maybeTarget string

	if user {
		maybeUser = "--user "
	}
	if target != "" {
		maybeTarget = fmt.Sprintf(" %s.%s", target, unitType)
	}

	return fmt.Sprintf("systemctl %s%s%s", maybeUser, action, maybeTarget)
}

func check(s entity.Service, action systemdAction, unitType string) (string, bool, error) {
	var negated bool

	checkAction := action.checkCmd
	negated = strings.HasPrefix(checkAction, negationPrefix)
	if negated {
		checkAction = strings.TrimLeft(checkAction, negationPrefix)
	}

	return systemctlCmd(checkAction, s.Name, unitType, !s.System), negated, nil
}

func actuate(s entity.Service, action systemdAction, unitType string) (string, error) {
	return systemctlCmd(action.actuateCmd, s.Name, unitType, !s.System), nil
}

func ensureServiceState(s entity.Service, actionStr, unitType string) error {
	action, ok := actions[actionStr]
	if !ok {
		return fmt.Errorf("no such action: %s", actionStr)
	}

	checkCmd, negated, err := check(s, action, unitType)
	if err != nil {
		return err
	}

	// Don't need sudo for check actions, so don't use runSystemctlCmd
	resp, _ := marecmd.RunFmtErr(marecmd.Input{Command: checkCmd})
	if negated && resp.Code != 0 {
		return nil
	} else if !negated && resp.Code == 0 {
		return nil
	}

	actuateCmd, err := actuate(s, action, unitType)
	if err != nil {
		return err
	}

	caser := cases.Title(language.Und)
	verb := caser.String(action.logVerb)
	internal.Logger.Debug().Str("name", s.Name).Str("state", verb).Msg("Ensuring service state")

	return runSystemctlCmd(actuateCmd, s)
}

func enable(s entity.Service) error {
	if s.DontEnable {
		return nil
	}

	if s.Unit != nil && s.Unit.Type != oneshotService {
		err := ensureServiceState(s, "enable", "service")
		if err != nil {
			return err
		}
	}

	if s.Timer != nil {
		return ensureServiceState(s, "enable", "timer")
	}

	return nil
}

func start(s entity.Service) error {
	if s.DontStart {
		return nil
	}

	if s.Unit != nil && s.Unit.Type != oneshotService {
		err := ensureServiceState(s, "start", "service")
		if err != nil {
			return err
		}
	}

	if s.Timer != nil {
		return ensureServiceState(s, "start", "timer")
	}

	return nil
}

func expandService(s entity.Service, cfg entity.Config) (entity.Service, error) {
	if s.DontTemplate || s.Unit == nil {
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

func maybeStop(s entity.Service) error {
	if !s.Stop {
		return nil
	}
	err := ensureServiceState(s, "stop", "service")
	if err != nil {
		internal.Logger.Error().Err(err).Str("name", s.Name).Msg("Error ensuring service state")
		return err
	}

	if s.Timer != nil {
		err = ensureServiceState(s, "stop", "timer")
		if err != nil {
			internal.Logger.Error().Err(err).Str("name", s.Name).Msg("Error stopping timer")
			return err
		}
	}

	return nil
}

func maybeDisable(s entity.Service) error {
	if !s.Disable {
		return nil
	}
	err := ensureServiceState(s, "disable", "service")
	if err != nil {
		internal.Logger.Error().Err(err).Str("name", s.Name).Msg("Error disabling service")
		return err
	}

	if s.Timer != nil {
		err = ensureServiceState(s, "disable", "timer")
		if err != nil {
			internal.Logger.Error().Err(err).Str("name", s.Name).Msg("Error disabling timer")
			return err
		}
	}

	return nil
}

func initService(s entity.Service, cfg entity.Config) error {
	if !when.ShouldRun(s) {
		whenText := internal.PrettyLogStr(s.When)
		internal.Logger.Trace().Str("name", s.Name).Str("when", whenText).Msg(
			"Skipping service initialization")
		return nil
	}

	err := maybeStop(s)
	if err != nil {
		return err
	}

	err = maybeDisable(s)
	if err != nil {
		return err
	}

	name := s.Name
	s, err = expandService(s, cfg)
	if err != nil {
		internal.Logger.Error().Str("name", name).Err(err).Msg("Error expanding service")
		return err
	}

	err = persist(s)
	if err != nil {
		internal.Logger.Error().Str("name", name).Err(err).Msg("Error persisting service")
		return err
	}

	err = enable(s)
	if err != nil {
		internal.Logger.Error().Str("name", name).Err(err).Msg("Error enabling service")
		return err
	}

	err = start(s)
	if err != nil {
		internal.Logger.Error().Str("name", name).Err(err).Msg("Error starting service")
		return err
	}

	return nil
}

func initServices(config entity.Config) error {
	var svcErr []error
	for _, svc := range config.Services {
		err := initService(svc, config)
		svcErr = append(svcErr, err)
	}

	return errors.Join(svcErr...)
}
