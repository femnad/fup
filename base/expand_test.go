package base

import (
	"testing"
)

func TestExpandSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
		input    string
		expanded string
	}{
		{
			name:     "Test simple expansion",
			settings: Settings{ExtractDir: "foo"},
			input:    "${extract_dir}/bar",
			expanded: "foo/bar",
		},
		{
			name:     "Test multi expansion",
			settings: Settings{ExtractDir: "foo", CloneDir: "baz"},
			input:    "${extract_dir}/bar/${clone_dir}",
			expanded: "foo/bar/baz",
		},
		{
			name:     "Test non expandable left intact",
			settings: Settings{ExtractDir: "foo", CloneDir: "baz"},
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

func TestExpandWithLookup(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
		input    string
		expanded string
		lookup   map[string]string
	}{
		{
			name:     "Test simple lookup",
			settings: Settings{},
			input:    "${baz}/bar",
			expanded: "foo/bar",
			lookup:   map[string]string{"baz": "foo"},
		},
		{
			name:     "Test settings and lookup expansion",
			settings: Settings{CloneDir: "baz"},
			input:    "${baz}/bar/${clone_dir}",
			expanded: "fred/bar/baz",
			lookup:   map[string]string{"baz": "fred"},
		},
		{
			name:     "Test override with lookup",
			settings: Settings{ExtractDir: "foo", CloneDir: "baz"},
			input:    "${extract_dir}/bar/${clone_dir}",
			expanded: "xyz/bar/baz",
			lookup:   map[string]string{"extract_dir": "xyz"},
		},
		{
			name:     "Non-resolved left alone",
			settings: Settings{ExtractDir: "foo", CloneDir: "baz"},
			input:    "${extract_dir}/bar/${clone_dir}/qux/${test_dir}/test/${var}",
			expanded: "foo/bar/baz/qux/xyz/test/${var}",
			lookup:   map[string]string{"test_dir": "xyz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expanded := ExpandSettingsWithLookup(tt.settings, tt.input, tt.lookup)
			if expanded != tt.expanded {
				t.Errorf("Expected expanded string to be %s, but it was %s", tt.expanded, expanded)
			}
		})
	}
}
