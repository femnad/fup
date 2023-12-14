package unless

import (
	"testing"
)

func Test_applyProc(t *testing.T) {
	type args struct {
		proc   string
		output string
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
				proc:   "split 0",
				output: "qux bar baz",
			},
			want: "qux",
		},
		{
			name: "Split by comma then space",
			args: args{
				proc:   `splitBy "," 0 | split 1`,
				output: "qux bar baz, fred",
			},
			want: "bar",
		},
		{
			name: "Split negative index",
			args: args{
				proc:   "split -1",
				output: "qux bar baz",
			},
			want: "baz",
		},
		{
			name: "Split invalid index",
			args: args{
				proc:   "split 3",
				output: "qux bar baz",
			},
			wantErr: true,
		},
		{
			name: "Split invalid negative index",
			args: args{
				proc:   "split -3",
				output: "qux bar",
			},
			wantErr: true,
		},
		{
			name: "Cut from index",
			args: args{
				proc:   "cut 1",
				output: "v123",
			},
			want: "123",
		},
		{
			name: "Cut from negative index",
			args: args{
				proc:   "cut -2",
				output: "v123.345",
			},
			want: "45",
		},
		{
			name: "Head first",
			args: args{
				proc: "head 0",
				output: `bar
baz
qux`,
			},
			want: "bar",
		},
		{
			name: "Head last",
			args: args{
				proc: "head -1",
				output: `bar
baz
qux`,
			},
			want: "qux",
		},
		{
			name: "Head split cut",
			args: args{
				proc: "head 0 | split -1 | cut 1",
				output: `foo v1.2.3
baz`,
			},
			want: "1.2.3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyProc(tt.args.proc, tt.args.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyProc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("applyProc() got = %v, want %v", got, tt.want)
			}
		})
	}
}
