package base

import (
	"testing"

	"github.com/femnad/fup/base/settings"
)

func TestExpandSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings settings.Settings
		input    string
		expanded string
	}{
		{
			name:     "Test simple expansion",
			settings: settings.Settings{ExtractDir: "foo"},
			input:    "${extract_dir}/bar",
			expanded: "foo/bar",
		},
		{
			name:     "Test multi expansion",
			settings: settings.Settings{ExtractDir: "foo", CloneDir: "baz"},
			input:    "${extract_dir}/bar/${clone_dir}",
			expanded: "foo/bar/baz",
		},
		{
			name:     "Test non expandable left intact",
			settings: settings.Settings{ExtractDir: "foo", CloneDir: "baz"},
			input:    "${extract_dir}/bar/${clone_dir}/qux/${some_var}",
			expanded: "foo/bar/baz/qux/${some_var}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expanded := ExpandSettings(tt.settings, tt.input)
			if expanded != tt.expanded {
				t.Errorf("Expected expanded string to be %s, but it was %s", tt.expanded, expanded)
			}
		})
	}
}
