package common

import (
	"reflect"
	"testing"
)

func TestRightSplit(t *testing.T) {
	type args struct {
		s   string
		sep string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "String with one separator",
			args: args{
				s:   "x.y",
				sep: ".",
			},
			want: []string{"x", "y"},
		},
		{
			name: "String with two separators",
			args: args{
				s:   "x.y.z",
				sep: ".",
			},
			want: []string{"x.y", "z"},
		},
		{
			name: "String with no separator",
			args: args{
				s:   "xy",
				sep: ".",
			},
			want: []string{"xy"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RightSplit(tt.args.s, tt.args.sep); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RightSplit() = %v, want %v", got, tt.want)
			}
		})
	}
}
