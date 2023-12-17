package internal

import (
	"testing"
)

func Test_RunTemplateFn(t *testing.T) {
	type args struct {
		proc  string
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Simple split",
			args: args{
				proc:  "split 0",
				input: "qux bar baz",
			},
			want: "qux",
		},
		{
			name: "Split by comma then space",
			args: args{
				proc:  `splitBy "," 0 | split 1`,
				input: "qux bar baz, fred",
			},
			want: "bar",
		},
		{
			name: "Split negative index",
			args: args{
				proc:  "split -1",
				input: "qux bar baz",
			},
			want: "baz",
		},
		{
			name: "Split invalid index",
			args: args{
				proc:  "split 3",
				input: "qux bar baz",
			},
			wantErr: true,
		},
		{
			name: "Split invalid negative index",
			args: args{
				proc:  "split -3",
				input: "qux bar",
			},
			wantErr: true,
		},
		{
			name: "Cut from index",
			args: args{
				proc:  "cut 1",
				input: "v123",
			},
			want: "123",
		},
		{
			name: "Cut from negative index",
			args: args{
				proc:  "cut -2",
				input: "v123.345",
			},
			want: "45",
		},
		{
			name: "Head first",
			args: args{
				proc: "head 0",
				input: `bar
baz
qux`,
			},
			want: "bar",
		},
		{
			name: "Head last",
			args: args{
				proc: "head -1",
				input: `bar
baz
qux`,
			},
			want: "qux",
		},
		{
			name: "Head split cut",
			args: args{
				proc: "head 0 | split -1 | cut 1",
				input: `foo v1.2.3
baz`,
			},
			want: "1.2.3",
		},
		{
			name: "Reverse cut",
			args: args{
				proc:  "revCut 1",
				input: "v123",
			},
			want: "v",
		},
		{
			name: "Reverse cut with negative index",
			args: args{
				proc:  "revCut -4",
				input: "v123.345",
			},
			want: "v123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunTemplateFn(tt.args.input, tt.args.proc)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunTemplateFn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RunTemplateFn() got = %v, want %v", got, tt.want)
			}
		})
	}
}
