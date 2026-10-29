package config

import (
	"fmt"
	"os"
	"strconv"
)

// ShellPID identifies the owning shell even when a wrapper uses command substitution.
func ShellPID() (int, error) {
	value := os.Getenv("KONF_SHELL_PID")
	if value == "" {
		return os.Getppid(), nil
	}
	pid, err := strconv.Atoi(value)
	if err != nil || pid <= 0 {
		return 0, fmt.Errorf("invalid KONF_SHELL_PID %q: expected a positive process ID", value)
	}
	return pid, nil
}

var curConf *Config = &Config{}

// Config describes all values that can currently be configured for konf
type Config struct {
	KonfDir string
	Silent  bool
}

// DefaultConfig returns an initialized config based on the users HomeDir
func DefaultConfig() (*Config, error) {
	c := &Config{}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	c.KonfDir = home + "/.kube/konfs"
	c.Silent = false

	return c, nil
}

// SetGlobalConfig sets the config to the config supplied as its argument
func SetGlobalConfig(or *Config) {
	curConf = or
}

// Currently there is no need to customize store and active configs individually.
// Setting the konfDir should be enough

// ActiveDir returns the currently configured active directory
func ActiveDir() string {
	return curConf.KonfDir + "/active"
}

// StoreDir returns the currently configured store directory
func StoreDir() string {
	return curConf.KonfDir + "/store"
}

// LatestKonfFilePath returns the currently configured latest konf file
func LatestKonfFilePath() string {
	return curConf.KonfDir + "/latestkonf"
}
