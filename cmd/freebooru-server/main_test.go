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

func TestHTTPAddress(t *testing.T) {
	tests := []struct {
		name    string
		app     fakeConfigLoader
		want    string
		wantErr bool
	}{
		{
			name: "configured address",
			app: fakeConfigLoader{
				appConfig: config.AppConfig{HTTPAddress: "::1", HTTPPort: 53123},
			},
			want: "[::1]:53123",
		},
		{
			name: "default port after load failure",
			app: fakeConfigLoader{
				loadErr: errors.New("configuration is unavailable"),
			},
			want:    "0.0.0.0:52800",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := httpAddress(context.Background(), test.app)
			if got != test.want {
				t.Fatalf("httpAddress() = %q, want %q", got, test.want)
			}
			if (err != nil) != test.wantErr {
				t.Fatalf("httpAddress() error = %v, want error %t", err, test.wantErr)
			}
		})
	}
}
