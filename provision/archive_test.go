package provision

import (
	"io/fs"
	"reflect"
	"testing"
	"time"

	"github.com/femnad/fup/entity"
)

type mockFileInfo struct {
	mode  uint32
	name  string
	isDir bool
}

func (m mockFileInfo) Name() string {
	//TODO implement me
	panic("implement me")
}

func (m mockFileInfo) Size() int64 {
	//TODO implement me
	panic("implement me")
}

func (m mockFileInfo) Mode() fs.FileMode {
	return fs.FileMode(m.mode)
}

func (m mockFileInfo) ModTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (m mockFileInfo) IsDir() bool {
	return m.isDir
}

func (m mockFileInfo) Sys() any {
	//TODO implement me
	panic("implement me")
}

func mockExec(name string) mockFileInfo {
	return mockFileInfo{mode: 493, name: name, isDir: false}
}

func mockFile(name string) mockFileInfo {
	return mockFileInfo{mode: 420, name: name, isDir: false}
}

func mockDir(name string) mockFileInfo {
	return mockFileInfo{mode: 493, name: name, isDir: true}
}

func Test_determineArchiveRoot(t *testing.T) {
	type args struct {
		archive entity.Release
		entries []archiveEntry
	}
	tests := []struct {
		name    string
		args    args
		want    ReleaseInfo
		wantErr bool
	}{
		{
			name: "Single file in a dir",
			args: args{
				entries: []archiveEntry{{
					info: mockExec("foo"),
					name: "dir/foo",
				}},
			},
			want: ReleaseInfo{
				hasRootDir: true,
				maybeExec:  "dir/foo",
				target:     "dir",
			},
		},
		{
			name: "Single file not in a dir",
			args: args{
				entries: []archiveEntry{{
					info: mockExec("foo"),
					name: "foo",
				}},
			},
			want: ReleaseInfo{
				hasRootDir: false,
				maybeExec:  "foo",
				target:     "foo",
			},
		},
		{
			name: "Multiple files without root dir with archive name",
			args: args{
				archive: entity.Release{Ref: "qux"},
				entries: []archiveEntry{
					{
						info: mockExec("foo"),
						name: "foo",
					},
					{
						info: mockDir("dir"),
						name: "dir",
					},
					{
						info: mockFile("baz"),
						name: "dir/baz",
					},
				},
			},
			want: ReleaseInfo{
				hasRootDir: false,
				maybeExec:  "foo",
				target:     "qux",
			},
		},
		{
			name: "Multiple files without root dir and archive name",
			args: args{
				entries: []archiveEntry{
					{
						info: mockExec("foo"),
						name: "foo",
					},
					{
						info: mockDir("dir"),
						name: "dir",
					},
					{
						info: mockFile("baz"),
						name: "dir/baz",
					},
				},
			},
			want: ReleaseInfo{
				hasRootDir: false,
				maybeExec:  "foo",
				target:     "foo",
			},
		},
		{
			name: "Multiple files with root dir",
			args: args{
				entries: []archiveEntry{
					{
						info: mockExec("foo"),
						name: "qux/baz/foo",
					},
					{
						info: mockDir("fred"),
						name: "qux/fred",
					},
					{
						info: mockFile("baz"),
						name: "qux/bar/baz",
					},
				},
			},
			want: ReleaseInfo{
				hasRootDir: true,
				maybeExec:  "qux/baz/foo",
				target:     "qux",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getReleaseInfo(tt.args.archive, tt.args.entries)
			if (err != nil) != tt.wantErr {
				t.Errorf("determineArchiveRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("determineArchiveRoot() got = %v, want %v", got, tt.want)
			}
		})
	}
}
