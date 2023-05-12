package provision

import (
	"github.com/femnad/fup/base"
	"testing"
)

func Test_writeTmpl(t *testing.T) {
	type args struct {
		s base.Service
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Basic service",
			args: args{s: base.Service{Unit: base.Unit{Desc: "Test", Exec: "test"}}},
			want: `[Unit]
Description=Test

[Service]
ExecStart=test

[Install]
WantedBy=default.target
`,
		},
		{
			name: "Service with options",
			args: args{s: base.Service{Unit: base.Unit{Desc: "Test", Exec: "test",
				Options: map[string]string{"Restart": "always", "RestartSec": "5"}}}},
			want: `[Unit]
Description=Test

[Service]
ExecStart=test
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := writeTmpl(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeTmpl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			out := got.buf.String()
			if out != tt.want {
				t.Errorf("writeTmpl() got = \n`%s`, want \n`%s`", out, tt.want)
			}
		})
	}
}
