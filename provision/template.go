package provision

import (
	"bytes"
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

func getTemplateContent(config base.Config, tmplSrc string) ([]byte, error) {
	if !config.IsRemote() {
		tmplPath := path.Join(internal.ExpandUser(config.Settings.TemplateDir), tmplSrc)
		return os.ReadFile(tmplPath)
	}

	configBase, _ := path.Split(config.Url())
	_, relTmplDir := path.Split(config.Settings.TemplateDir)
	tmplUrl, err := url.JoinPath(configBase, relTmplDir, tmplSrc)
	if err != nil {
		return nil, err
	}

	return remote.ReadResponseBytes(tmplUrl)
}

func applyTemplate(tmpl base.Template, config base.Config) error {
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
