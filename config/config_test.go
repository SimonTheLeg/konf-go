package config

import (
	"os"
	"testing"
)

func TestShellPID(t *testing.T) {
	for _, value := range []string{"", "42", "0", "-1", "invalid", "99999999999999999999999999999"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("KONF_SHELL_PID", value)
			pid, err := ShellPID()
			switch value {
			case "":
				if err != nil || pid != os.Getppid() {
					t.Fatalf("got %d, %v; want parent %d", pid, err, os.Getppid())
				}
			case "42":
				if err != nil || pid != 42 {
					t.Fatalf("got %d, %v; want 42", pid, err)
				}
			default:
				if err == nil {
					t.Fatalf("accepted invalid PID %q", value)
				}
			}
		})
	}
}
