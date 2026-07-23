package main

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"
)

func TestServerCommandsDelegateToUserSystemd(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want [][]string
	}{
		{name: "start", args: []string{"server", "start"}, want: [][]string{{"systemctl", "--user", "daemon-reload"}, {"systemctl", "--user", "enable", "--now", serverService}}},
		{name: "stop", args: []string{"server", "stop"}, want: [][]string{{"systemctl", "--user", "disable", "--now", serverService}}},
		{name: "restart", args: []string{"server", "restart"}, want: [][]string{{"systemctl", "--user", "restart", serverService}}},
		{name: "status", args: []string{"server", "status"}, want: [][]string{{"systemctl", "--user", "status", serverService}}},
		{name: "logs", args: []string{"server", "logs"}, want: [][]string{{"journalctl", "--user", "--unit", serverService}}},
		{name: "follow logs", args: []string{"server", "logs", "--follow"}, want: [][]string{{"journalctl", "--user", "--unit", serverService, "--follow"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got [][]string
			options := &rootOptions{newCore: defaultCoreFactory}
			options.runCommand = func(
				_ context.Context,
				_ io.Reader,
				_, _ io.Writer,
				name string,
				args ...string,
			) error {
				got = append(got, append([]string{name}, args...))
				return nil
			}
			command := newRootCommandWithOptions(options)
			command.SetOut(new(bytes.Buffer))
			command.SetErr(new(bytes.Buffer))
			command.SetArgs(test.args)
			if err := command.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("command = %#v, want %#v", got, test.want)
			}
		})
	}
}
