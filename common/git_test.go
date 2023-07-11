package common

import (
	"reflect"
	"testing"
)

func Test_processUrl(t *testing.T) {
	type args struct {
		repoUrl string
	}
	tests := []struct {
		name    string
		args    args
		want    cloneRef
		wantErr bool
	}{
		{
			name: "URL with scheme and host",
			args: args{repoUrl: "https://github.com/foo/bar.git"},
			want: cloneRef{
				url:  "https://github.com/foo/bar.git",
				base: "bar",
			},
		},
		{
			name: "URL with scheme and host, no git suffix ",
			args: args{repoUrl: "https://github.com/foo/bar"},
			want: cloneRef{
				url:  "https://github.com/foo/bar.git",
				base: "bar",
			},
		},
		{
			name: "URL with no scheme",
			args: args{repoUrl: "github.com/foo/bar"},
			want: cloneRef{
				url:  "git@github.com:foo/bar.git",
				base: "bar",
			},
		},
		{
			name: "URL with no host",
			args: args{repoUrl: "foo/bar"},
			want: cloneRef{
				url:  "git@github.com:foo/bar.git",
				base: "bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processUrl(tt.args.repoUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("processUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processUrl() got = %v, want %v", got, tt.want)
			}
		})
	}
}
