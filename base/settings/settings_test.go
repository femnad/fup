package settings

import "testing"

func Test_expand(t *testing.T) {
	type args struct {
		s      string
		lookup map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expand(tt.args.s, tt.args.lookup)
			if (err != nil) != tt.wantErr {
				t.Errorf("expand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("expand() got = %v, want %v", got, tt.want)
			}
		})
	}
}
