package provision

import (
	"testing"

	"github.com/femnad/fup/entity"
)

func Test_writeTmpl(t *testing.T) {
	type args struct {
		s entity.Service
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Basic service",
			args: args{s: entity.Service{
				Unit: entity.Unit{
					Desc: "Test",
					Exec: "test"},
			},
			},
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
			args: args{s: entity.Service{
				Unit: entity.Unit{
					Desc: "Test",
					Exec: "test",
					Options: map[string]string{
						"Restart":    "always",
						"RestartSec": "5"},
				},
			},
			},
			want: `[Unit]
Description=Test

[Service]
ExecStart=test
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`,
		},
		{
			name: "Service with wanted by",
			args: args{s: entity.Service{Unit: entity.Unit{
				Desc:     "Test",
				Exec:     "test",
				WantedBy: "sleep",
			}}},
			want: `[Unit]
Description=Test

[Service]
ExecStart=test

[Install]
WantedBy=sleep.target
`,
		},
		{
			name: "Service with before",
			args: args{s: entity.Service{Unit: entity.Unit{
				Desc:   "Test",
				Exec:   "test",
				Before: "sleep",
			}}},
			want: `[Unit]
Description=Test
Before=sleep.target

[Service]
ExecStart=test

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
			if got != tt.want {
				t.Errorf("writeTmpl() got = \n`%s`, want \n`%s`", got, tt.want)
			}
		})
	}
}
