package provision

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

const (
	// -1 means don't change user/group with os.Chown.
	defaultId           = -1
	getentUserDatabase  = "passwd"
	getentGroupDatabase = "group"
	getentSeparator     = ":"
	rootUser            = "root"
	tmpDir              = "/tmp"
)

func getent(key, database string) (int, error) {
	if key == "" {
		return defaultId, nil
	}

	out, err := common.RunCmd(common.CmdIn{Command: fmt.Sprintf("getent %s %s", database, key)})
	if err != nil {
		return 0, err
	}

	getentOutput := out.Stdout
	getentFields := strings.Split(getentOutput, getentSeparator)
	if len(getentFields) < 2 {
		return 0, fmt.Errorf("unexpected getent output: %s", getentOutput)
	}

	id, err := strconv.ParseInt(getentFields[2], 10, 32)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func chown(file, user, group string) error {
	isHomePath := common.IsHomePath(file)
	if user == "" && group == "" {
		if isHomePath {
			return nil
		}

		user = rootUser
		group = rootUser
	}

	userId, err := getent(user, getentUserDatabase)
	if err != nil {
		return err
	}

	groupId, err := getent(group, getentGroupDatabase)
	if err != nil {
		return err
	}

	if isHomePath {
		return os.Chown(file, userId, groupId)
	}

	chownCmd := "chown "
	if user != "" {
		chownCmd += user
	}
	if group != "" {
		chownCmd += ":" + user
	}
	chownCmd += " " + file

	out, err := common.RunCmd(common.CmdIn{Command: chownCmd, Sudo: true})
	if err != nil {
		return fmt.Errorf("error changing owner of %s, output %s: %v", file, out.Stderr, err)
	}

	return nil
}

func applyTemplate(tmpl base.Template, config base.Config) error {
	dstDir, dstFile := path.Split(internal.ExpandUser(tmpl.Dest))
	file, err := os.CreateTemp(tmpDir, dstFile)
	if err != nil {
		return err
	}
	defer file.Close()

	tmplPath := path.Join(internal.ExpandUser(config.Settings.TemplateDir), tmpl.Src)
	templateContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return err
	}
	parsed, err := template.New(dstFile).Parse(string(templateContent))
	if err != nil {
		return err
	}

	err = parsed.Execute(file, tmpl.Context)
	if err != nil {
		return err
	}

	src := file.Name()
	dst := tmpl.Dest

	mode := os.FileMode(tmpl.Mode)
	if mode != 0 {
		err = os.Chmod(src, os.FileMode(tmpl.Mode))
		if err != nil {
			return err
		}
	}

	srcSum, err := common.Checksum(src)
	if err != nil {
		return fmt.Errorf("error getting checksum of source %s: %v", src, err)
	}

	var dstSum string
	var dstNotExist bool
	_, err = os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			dstNotExist = true
		} else {
			return err
		}
	}

	if !dstNotExist {
		dstSum, err = common.Checksum(dst)
		if err != nil {
			return fmt.Errorf("error getting checksum of destination %s: %v", dst, err)
		}
	}

	if !dstNotExist && srcSum == dstSum {
		internal.Log.Debugf("%s and %s have the same content", tmplPath, dst)
		return os.Remove(src)
	}

	if err = common.EnsureDir(dstDir); err != nil {
		return err
	}

	mvCmd := common.GetMvCmd(src, dst)
	out, err := common.RunCmd(mvCmd)

	if err != nil {
		return fmt.Errorf("error moving file source %s dest %s output %s: %v", src, dst, out.Stderr, err)
	}

	return chown(dst, tmpl.User, tmpl.Group)
}

func applyTemplates(config base.Config) {
	for _, tmpl := range config.Templates {
		if !when.ShouldRun(tmpl) {
			continue
		}

		err := applyTemplate(tmpl, config)
		if err != nil {
			internal.Log.Errorf("error applying template: %v", err)
			continue
		}
	}
}
