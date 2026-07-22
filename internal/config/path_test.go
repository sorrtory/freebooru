package config

import (
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{name: "home variable", path: "$HOME/data", want: filepath.Join(home, "data")},
		{name: "home shorthand", path: "~/data", want: filepath.Join(home, "data")},
		{name: "absolute", path: "/srv/freebooru", want: "/srv/freebooru"},
		{name: "relative", path: "data", wantErr: true},
		{name: "another user", path: "~someone/data", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ExpandPath(%q) succeeded, want error", tt.path)
				}
				return
			}
			if err != nil {
				t.Fatalf("ExpandPath(%q) error = %v", tt.path, err)
			}
			if got != tt.want {
				t.Fatalf("ExpandPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
