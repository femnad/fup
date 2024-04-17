package internal

import (
	"reflect"
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
			name: "Split negative index smaller than -1",
			args: args{
				proc:  "split -2",
				input: "qux bar baz",
			},
			want: "bar",
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
		{
			name: "Simple tail",
			args: args{
				proc: "tail 0",
				input: `foo
bar
baz`,
			},
			want: "baz",
		},
		{
			name: "Negative tail",
			args: args{
				proc: "tail -1",
				input: `foo
bar
baz`,
			},
			want: "foo",
		},
		{
			name: "Tail error",
			args: args{
				proc: "tail 3",
				input: `bar
baz`,
			},
			wantErr: true,
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

func Test_iterMap(t *testing.T) {
	type args struct {
		items []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "Single pair",
			args: args{items: []string{"a", "b"}},
			want: map[string]string{"a": "b"},
		},
		{
			name: "Multiple pairs",
			args: args{items: []string{"a", "b", "c", "d"}},
			want: map[string]string{"a": "b", "c": "d"},
		},
		{
			name: "No items",
			args: args{items: []string{}},
			want: map[string]string{},
		},
		{
			name:    "Odd number of items",
			args:    args{items: []string{"a", "b", "c"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := iterMap(tt.args.items...)
			if (err != nil) != tt.wantErr {
				t.Errorf("iterMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("iterMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}
