package settings

import (
	"os"
	"testing"
)

func Test_expand(t *testing.T) {
	type args struct {
		s      string
		lookup map[string]string
		env    map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Whole string expand",
			args: args{
				s:      "${foo}",
				lookup: map[string]string{"foo": "bar"},
			},
			want: "bar",
		},
		{
			name: "Var within string expand",
			args: args{
				s:      "val=${baz}",
				lookup: map[string]string{"baz": "qux"},
			},
			want: "val=qux",
		},
		{
			name: "Multiple expand",
			args: args{
				s:      "echo val1=${foo} val2=${baz} 2>&1",
				lookup: map[string]string{"foo": "bar", "baz": "qux"},
			},
			want: "echo val1=bar val2=qux 2>&1",
		},
		{
			name: "Escape variable",
			args: args{
				s:      "echo \\${x}",
				lookup: map[string]string{"x": "fred"},
			},
			want: "echo ${x}",
		},
		{
			name: "Don't expand vars without braces",
			args: args{
				s:      "echo $x",
				lookup: map[string]string{"x": "fred"},
			},
			want: "echo $x",
		},
		{
			name: "Don't expand undefined vars",
			args: args{
				s:      "echo ${x}; sleep 1",
				lookup: map[string]string{},
			},
			want: "echo ${x}; sleep 1",
		},
		{
			name: "Expand environment vars",
			args: args{
				s:      "echo ${baz}",
				lookup: map[string]string{},
				env:    map[string]string{"baz": "foo"},
			},
			want: "echo foo",
		},
		{
			name: "Custom lookup has precedence over environment variables",
			args: args{
				s:      "echo ${baz}",
				lookup: map[string]string{"baz": "qux"},
				env:    map[string]string{"baz": "foo"},
			},
			want: "echo qux",
		},
		{
			name: "Handle non-existing environment variables",
			args: args{
				s:      "echo ${fred}",
				lookup: map[string]string{},
				env:    map[string]string{},
			},
			want: "echo ${fred}",
		},
		{
			name: "Keep backslashes intact if they're not preceding dollar signs",
			args: args{
				s:      "echo 'foo \\'bar baz\\''",
				lookup: map[string]string{},
				env:    map[string]string{},
			},
			want: "echo 'foo \\'bar baz\\''",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.args.env {
				err := os.Setenv(k, v)
				if err != nil {
					t.Errorf("error setting env in %s: %v", tt.name, err)
				}
			}
			got := expand(tt.args.s, tt.args.lookup)
			if got != tt.want {
				t.Errorf("expand() got = %v, want %v", got, tt.want)
			}
		})
	}
}
