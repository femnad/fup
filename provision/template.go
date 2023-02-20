package provision

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

const tmpDir = "/tmp"

func getMvCmd(src, dst string) string {
	home := os.Getenv("HOME")
	cmd := fmt.Sprintf("mv %s %s", src, dst)

	sudo := !strings.HasPrefix(dst, home)
	if sudo {
		return fmt.Sprintf("sudo %s", cmd)
	}

	return cmd
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

	srcSum, err := checksum(src)
	if err != nil {
		return fmt.Errorf("error getting checksum of source %s: %v", src, err)
	}
	dstSum, err := checksum(dst)
	if err != nil {
		return fmt.Errorf("error getting checksum of destination %s: %v", dst, err)
	}

	if srcSum == dstSum {
		internal.Log.Debugf("%s and %s have the same content", tmplPath, dst)
		return os.Remove(src)
	}

	_, err = os.Stat(dstDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dstDir, 0744)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	mvCmd := getMvCmd(src, dst)
	out, err := common.RunCmdGetStderr(mvCmd)

	if err != nil {
		return fmt.Errorf("error moving file source %s dest %s output %s: %v", src, dst, out, err)
	}

	return nil
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
