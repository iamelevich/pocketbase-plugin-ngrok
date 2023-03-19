package pocketbase_plugin_ngrok

import (
	"context"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func TestPlugin_Validate(t *testing.T) {
	type fields struct {
		app     core.App
		options *Options
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "App is nil",
			fields: fields{
				app: nil,
				options: &Options{
					Ctx:       context.Background(),
					Enabled:   true,
					AuthToken: "",
				},
			},
			wantErr: true,
		},
		{
			name: "Context is nil",
			fields: fields{
				app: pocketbase.New(),
				options: &Options{
					Ctx:       nil,
					Enabled:   true,
					AuthToken: "",
				},
			},
			wantErr: true,
		},
		{
			name: "Enabled but no AuthToken",
			fields: fields{
				app: pocketbase.New(),
				options: &Options{
					Ctx:       context.Background(),
					Enabled:   true,
					AuthToken: "",
				},
			},
			wantErr: true,
		},
		{
			name: "Not enabled and has no AuthToken",
			fields: fields{
				app: pocketbase.New(),
				options: &Options{
					Ctx:       context.Background(),
					Enabled:   false,
					AuthToken: "",
				},
			},
			wantErr: false,
		},
		{
			name: "Enabled and has AuthToken",
			fields: fields{
				app: pocketbase.New(),
				options: &Options{
					Ctx:       context.Background(),
					Enabled:   true,
					AuthToken: "TEST_AUTH_TOKEN",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{
				app:     tt.fields.app,
				options: tt.fields.options,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
