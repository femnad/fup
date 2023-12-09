package provision

import (
	"bytes"
	"errors"
	"net/url"
	"os"
	"path"
	"text/template"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
	"github.com/femnad/fup/remote"
)

const (
	tmpDir = "/tmp"
)

func getTemplateDir(config base.Config) (string, error) {
	templateDir := internal.ExpandUser(config.Settings.TemplateDir)
	if path.IsAbs(templateDir) {
		return templateDir, nil
	}

	configPath := config.File()
	configDir, _ := path.Split(configPath)
	if path.IsAbs(configDir) {
		return path.Join(configDir, templateDir), nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return path.Join(wd, configDir, templateDir), nil
}

func getLocalTemplate(config base.Config, tmplSrc string) ([]byte, error) {
	templateDir, err := getTemplateDir(config)
	if err != nil {
		return nil, err
	}

	tmplPath := path.Join(templateDir, tmplSrc)
	return os.ReadFile(tmplPath)
}

func getTemplateContent(config base.Config, tmplSrc string) ([]byte, error) {
	if !config.IsRemote() {
		return getLocalTemplate(config, tmplSrc)
	}

	configBase, _ := path.Split(config.File())
	_, relTmplDir := path.Split(config.Settings.TemplateDir)
	tmplUrl, err := url.JoinPath(configBase, relTmplDir, tmplSrc)
	if err != nil {
		return nil, err
	}

	return remote.ReadResponseBytes(tmplUrl)
}

func applyTemplate(tmpl base.Template, config base.Config) error {
	if !when.ShouldRun(tmpl) {
		return nil
	}

	templateContent, err := getTemplateContent(config, tmpl.Src)
	if err != nil {
		return err
	}

	parsed, err := template.New("tmpl").Parse(string(templateContent))
	if err != nil {
		return err
	}

	tmplBuffer := bytes.Buffer{}
	err = parsed.Execute(&tmplBuffer, tmpl.Context)
	if err != nil {
		return err
	}

	_, err = common.WriteContent(common.ManagedFile{Path: tmpl.Dest, Content: tmplBuffer.String()})
	if err != nil {
		return err
	}

	return nil
}

func applyTemplates(config base.Config) error {
	var tmplErr []error
	for _, tmpl := range config.Templates {
		err := applyTemplate(tmpl, config)
		if err != nil {
			internal.Log.Errorf("error applying template: %v", err)
		}
		tmplErr = append(tmplErr, err)
	}

	return errors.Join(tmplErr...)
}
