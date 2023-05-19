package provision

import (
	"bytes"
	"os"
	"path"
	"text/template"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

const (
	tmpDir = "/tmp"
)

func applyTemplate(tmpl base.Template, config base.Config) error {
	tmplPath := path.Join(internal.ExpandUser(config.Settings.TemplateDir), tmpl.Src)
	templateContent, err := os.ReadFile(tmplPath)
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
