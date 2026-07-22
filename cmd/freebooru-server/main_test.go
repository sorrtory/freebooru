package main

import (
	"context"
	"errors"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

type fakeConfigLoader struct {
	appConfig config.AppConfig
	loadErr   error
}

func (f fakeConfigLoader) LoadConfig(context.Context) error {
	return f.loadErr
}

func (f fakeConfigLoader) AppConfig() config.AppConfig {
	return f.appConfig
}

func TestHTTPPort(t *testing.T) {
	tests := []struct {
		name    string
		app     fakeConfigLoader
		want    int
		wantErr bool
	}{
		{
			name: "configured port",
			app: fakeConfigLoader{
				appConfig: config.AppConfig{HTTPPort: 53123},
			},
			want: 53123,
		},
		{
			name: "default port after load failure",
			app: fakeConfigLoader{
				loadErr: errors.New("configuration is unavailable"),
			},
			want:    config.DefaultAppConfig().HTTPPort,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := httpPort(context.Background(), test.app)
			if got != test.want {
				t.Fatalf("httpPort() = %d, want %d", got, test.want)
			}
			if (err != nil) != test.wantErr {
				t.Fatalf("httpPort() error = %v, want error %t", err, test.wantErr)
			}
		})
	}
}
