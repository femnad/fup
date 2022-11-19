package common

import "testing"

func TestNameFromRepo(t *testing.T) {
	type args struct {
		repo string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test successful parse",
			args: args{repo: "https://github.com/foo/bar.git"},
			want: "bar",
		},
		{
			name:    "Test parsing error",
			args:    args{repo: "quz/bar/baz"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NameFromRepo(tt.args.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("NameFromRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NameFromRepo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
