//go:build integration
// +build integration

package integration

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/settings"
)

var (
	absoluteArtifactsDir = internal.ExpandUser("~/out")
	relativeArtifactsDir = "out"
	absoluteDirs         = settings.Settings{
		BinDir:     path.Join(absoluteArtifactsDir, "bin"),
		ReleaseDir: path.Join(absoluteArtifactsDir, "ext"),
	}
	relativeDirs = settings.Settings{
		BinDir:     path.Join(relativeArtifactsDir, "bin"),
		ReleaseDir: path.Join(relativeArtifactsDir, "ext"),
	}
)

func ensureAbs(filePath, root string) (string, error) {
	if path.IsAbs(filePath) {
		return filePath, nil
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	return path.Join(root, filePath), nil
}

func cleanupReleasesTest(configFile string, dirs ...string) error {
	err := internal.EnsureFileAbsent(configFile)
	if err != nil {
		return err
	}

	for _, d := range dirs {
		err = internal.EnsureDirAbsent(d)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestReleases(t *testing.T) {
	type symlink struct {
		linkName string
		target   string
	}

	tests := []struct {
		name     string
		release  entity.Release
		symlink  symlink
		absolute bool
	}{
		{
			name: "tar_archive_no_root_dir",
			release: entity.Release{
				Ref: "foo",
				Url: "https://github.com/femnad/fup/releases/download/test-payload/release-no-root-dir.tar",
			},
			symlink: symlink{
				linkName: "bin/foo",
				target:   "ext/foo/foo",
			},
		},
		{
			name: "tar_archive_root_dir_different_than_exec",
			release: entity.Release{
				Ref: "foo",
				Url: "https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.tar",
			},
			symlink: symlink{
				linkName: "bin/foo",
				target:   "ext/foo-1.2.3-amd64/foo",
			},
		},
		{
			name: "tar_archive_root_dir_different_than_exec_override_target",
			release: entity.Release{
				Ref:    "foo",
				Target: "fred",
				Url:    "https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.tar",
			},
			symlink: symlink{
				linkName: "bin/foo",
				target:   "ext/fred/foo",
			},
		},
		{
			name: "tar_archive_root_dir_same_as_exec",
			release: entity.Release{
				Ref: "foo",
				Url: "https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.tar",
			},
			symlink: symlink{
				linkName: "bin/foo",
				target:   "ext/foo/foo",
			},
		},
		{
			name: "tar_archive_root_dir_same_as_exec_override_target",
			release: entity.Release{
				Ref:    "foo",
				Target: "qux",
				Url:    "https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.tar",
			},
			symlink: symlink{
				linkName: "bin/foo",
				target:   "ext/qux/foo",
			},
		},
		{
			name: "zip_archive_no_root_dir",
			release: entity.Release{
				Ref: "foo",
				Url: "https://github.com/femnad/fup/releases/download/test-payload/release-no-root-dir.zip",
			},
			symlink: symlink{
				linkName: "bin/foo",
				target:   "ext/foo/foo",
			},
		},
		{
			name: "zip_archive_root_dir_different_than_exec",
			release: entity.Release{
				Ref: "foo",
				Url: "https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.zip",
			},
			symlink: symlink{
				linkName: "bin/foo",
				target:   "ext/foo-1.2.3-amd64/foo",
			},
		},
		{
			name: "zip_archive_root_dir_different_than_exec_override_target",
			release: entity.Release{
				Ref:    "foo",
				Target: "baz",
				Url:    "https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.zip",
			},
			symlink: symlink{
				linkName: "bin/foo",
				target:   "ext/baz/foo",
			},
		},
		{
			name: "zip_archive_root_dir_same_as_exec",
			release: entity.Release{
				Ref: "foo",
				Url: "https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.zip",
			},
			symlink: symlink{
				linkName: "bin/foo",
				target:   "ext/foo/foo",
			},
		},
		{
			name: "zip_archive_root_dir_same_as_exec_override_target",
			release: entity.Release{
				Ref:    "foo",
				Target: "bar",
				Url:    "https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.zip",
			},
			symlink: symlink{
				linkName: "bin/foo",
				target:   "ext/bar/foo",
			},
		},
		{
			name: "zip_archive_root_dir_same_as_exec_abs_dirs",
			release: entity.Release{
				Ref: "foo",
				Url: "https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.zip",
			},
			symlink: symlink{
				linkName: "~/out/bin/foo",
				target:   "~/out/ext/foo/foo",
			},
			absolute: true,
		},
		{
			name: "binary_file",
			release: entity.Release{
				Ref:    "baz",
				Target: "foo",
				Url:    "https://github.com/femnad/fup/releases/download/test-payload/release-binary",
			},
			symlink: symlink{
				linkName: "~/out/bin/baz",
				target:   "~/out/ext/foo/baz",
			},
			absolute: true,
		},
		{
			name: "binary_file_no_name",
			release: entity.Release{
				Target: "foo",
				Url:    "https://github.com/femnad/fup/releases/download/test-payload/release-binary",
			},
			symlink: symlink{
				linkName: "~/out/bin/release-binary",
				target:   "~/out/ext/foo/release-binary",
			},
			absolute: true,
		},
		{
			name: "binary_file_no_name_and_target",
			release: entity.Release{
				Ref: "",
				Url: "https://github.com/femnad/fup/releases/download/test-payload/release-binary",
			},
			symlink: symlink{
				linkName: "~/out/bin/release-binary",
				target:   "~/out/ext/release-binary/release-binary",
			},
			absolute: true,
		},
	}

	configFile := "fup.yml"
	defer func() {
		err := cleanupReleasesTest(configFile, absoluteArtifactsDir, relativeArtifactsDir)
		if err != nil {
			log.Fatalf("Error cleaing up test artifacts: %v", err)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifactsDir := relativeArtifactsDir
			stg := &relativeDirs
			if tt.absolute {
				artifactsDir = absoluteArtifactsDir
				stg = &absoluteDirs
			}

			cfg := entity.Config{Releases: []entity.Release{tt.release}, Settings: *stg}
			err := writeConfig(cfg, configFile)
			if err != nil {
				t.Errorf("Error writing config file %s: %v", configFile, err)
				return
			}

			linkName := internal.ExpandUser(tt.symlink.linkName)
			linkName, err = ensureAbs(linkName, artifactsDir)
			if err != nil {
				t.Fatalf("Error ensuring %s is absolute: %v", linkName, err)
				return
			}

			target := internal.ExpandUser(tt.symlink.target)
			target, err = ensureAbs(target, artifactsDir)
			if err != nil {
				t.Fatalf("Error ensuring %s is absolute: %v", target, err)
				return
			}

			err = runFup("release", configFile)
			if err != nil {
				t.Errorf("error running fup: %v", err)
				return
			}

			var targetLink string
			targetLink, err = os.Readlink(linkName)
			if err != nil {
				t.Errorf("Error resolving symlink %s: %v", target, err)
				return
			}

			if targetLink != target {
				t.Fatalf("Link %s expected to point to %s but points to %s", linkName, target, targetLink)
				return
			}

			err = cleanupReleasesTest(configFile, absoluteArtifactsDir, relativeArtifactsDir)
			if err != nil {
				t.Fatalf("Error cleaning up test artifacts: %v", err)
				return
			}
		})

	}
}
