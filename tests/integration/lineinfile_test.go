//go:build integration
// +build integration

package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/mare"
	marecmd "github.com/femnad/mare/cmd"
)

const (
	testDir = "lineinfile"
)

func TestHelloWorld(t *testing.T) {
	bazContent := `foo
bar
baz
`
	bazFile := fmt.Sprintf("%s/baz", testDir)

	tests := []struct {
		name       string
		files      map[string]string
		want       map[string]string
		lineInFile entity.LineInFile
	}{
		{
			name: "Exact match",
			files: map[string]string{
				"baz": bazContent,
			},
			want: map[string]string{"baz": `foo
qux
baz
`},
			lineInFile: entity.LineInFile{
				File: bazFile,
				Name: "replace",
				Replace: []entity.Replacement{{
					Old: "bar",
					New: "qux",
				}},
			},
		},
		{
			name: "No matches",
			files: map[string]string{
				"baz": bazContent,
			},
			want: map[string]string{"baz": `foo
bar
baz
`},
			lineInFile: entity.LineInFile{
				File: bazFile,
				Name: "replace",
				Replace: []entity.Replacement{{
					Old: "qux",
					New: "fred",
				}},
			},
		},
		{
			name: "Regex match",
			files: map[string]string{
				"baz": `foo
fred123
baz
`},
			want: map[string]string{"baz": `foo
barney
baz
`},
			lineInFile: entity.LineInFile{
				File: bazFile,
				Name: "replace",
				Replace: []entity.Replacement{{
					Old:   "fred[0-9]+",
					New:   "barney",
					Regex: true,
				}},
			},
		},
		{
			name: "Remove line",
			files: map[string]string{
				"baz": bazContent,
			},
			want: map[string]string{"baz": `foo
baz
`},
			lineInFile: entity.LineInFile{
				File: bazFile,
				Name: "replace",
				Replace: []entity.Replacement{{
					Old:    "bar",
					Absent: true,
				}},
			},
		},
		{
			name: "Ensure line",
			files: map[string]string{
				"baz": bazContent,
			},
			want: map[string]string{"baz": `foo
baz
`},
			lineInFile: entity.LineInFile{
				File: bazFile,
				Name: "replace",
				Replace: []entity.Replacement{{
					Old:    "bar",
					Absent: true,
				}},
			},
		},
		{
			name: "Multiple changes",
			files: map[string]string{
				"baz": `foo
bar74dc
baz`,
			},
			want: map[string]string{"baz": `qux
barney
nope
`},
			lineInFile: entity.LineInFile{
				File: bazFile,
				Name: "replace",
				Replace: []entity.Replacement{
					{
						Old: "foo",
						New: "qux",
					},
					{
						Old:   "bar[0-9]{2}[a-f]{2}",
						New:   "barney",
						Regex: true,
					},
					{
						Old:    "baz",
						Absent: true,
					},
					{
						Old:    "hello",
						New:    "nope",
						Ensure: true,
					},
				},
			},
		},
	}

	configFile := "lineinfile/fup.yml"
	for _, tt := range tests {
		cfg := entity.Config{EnsureLines: []entity.LineInFile{tt.lineInFile}}
		out, err := yaml.Marshal(cfg)
		if err != nil {
			t.Errorf("Error marshalling config: %v", err)
			return
		}

		configDir, _ := path.Split(configFile)
		err = mare.EnsureDir(configDir)
		if err != nil {
			t.Errorf("Error ensuring dir %s: %v", configDir, err)
			return
		}

		err = os.WriteFile(configFile, out, 0o600)
		if err != nil {
			t.Errorf("Error writing config file %s: %v", configFile, err)
			return
		}

		t.Run(tt.name, func(t *testing.T) {
			for file, content := range tt.files {
				err = mare.EnsureDir(testDir)
				if err != nil {
					t.Errorf("Error ensuring dir %s: %v", testDir, err)
					return
				}

				file = path.Join(testDir, file)
				err = os.WriteFile(file, []byte(content), 0o600)
				if err != nil {
					t.Errorf("error writing file content for %s: %v", file, err)
					return
				}
			}

			goPath := os.Getenv("GOPATH")
			if goPath == "" {
				goPath = internal.ExpandUser("~/go")
			}
			fup := path.Join(goPath, "bin", "fup")

			err = marecmd.RunErrOnly(marecmd.Input{Command: fmt.Sprintf("%s -p line -f %s", fup, configFile)})
			if err != nil {
				t.Errorf("error running fup: %v", err)
				return
			}

			var got []byte
			for file, wantContent := range tt.want {
				file = path.Join(testDir, file)
				got, err = os.ReadFile(file)
				if err != nil {
					t.Errorf("error reading file %s: %v", file, err)
					return
				}

				gotContent := string(got)
				if wantContent != gotContent {
					t.Errorf("Wanted `%s`, got `%s`", wantContent, gotContent)
					return
				}
			}
		})

		err = os.RemoveAll(testDir)
		if err != nil {
			t.Errorf("error removing test dir %s: %v", testDir, err)
		}
	}
}
