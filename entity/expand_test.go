package entity

import (
	"testing"

	"github.com/femnad/fup/settings"
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
			settings: settings.Settings{ReleaseDir: "foo"},
			input:    "${release_dir}/bar",
			expanded: "foo/bar",
		},
		{
			name:     "Test multi expansion",
			settings: settings.Settings{ReleaseDir: "foo", CloneDir: "baz"},
			input:    "${release_dir}/bar/${clone_dir}",
			expanded: "foo/bar/baz",
		},
		{
			name:     "Test non expandable left intact",
			settings: settings.Settings{ReleaseDir: "foo", CloneDir: "baz"},
			input:    "${release_dir}/bar/${clone_dir}/qux/${some_var}",
			expanded: "foo/bar/baz/qux/${some_var}",
		},
		{
			name:     "Don't expand if dollar sign is followed by parentheses",
			settings: settings.Settings{},
			input:    "$(rpm)",
			expanded: "$(rpm)",
		},
		{
			name:     "Dollar sign followed by parentheses inside surrounded by strings",
			settings: settings.Settings{},
			input:    "install -y https://example.com/foo-$(rpm -E %fedora) --verbose",
			expanded: "install -y https://example.com/foo-$(rpm -E %fedora) --verbose",
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
