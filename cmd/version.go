package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

type versionInfo struct {
	GitVersion string
	GitCommit  string
	BuildDate  string
	GoVersion  string
	Platform   string
	Compiler   string
}

// defaultVersion will be returned if no ldflags were provided e.g. when running
// go build
var defaultVersionInfo versionInfo = versionInfo{
	GitVersion: "dev",
	GitCommit:  "dev",
	BuildDate:  "1970-01-01T00:00:00Z",
	GoVersion:  runtime.Version(),
	Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	Compiler:   runtime.Compiler,
}

type versionCmd struct {
	cmd *cobra.Command
}

func newVersionCommand() *versionCmd {
	vc := &versionCmd{}
	vc.cmd = &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Long:  "Print version and build info in a json format",
		RunE:  vc.version,
		Args:  cobra.ExactArgs(0),
	}

	return vc
}

func (c *versionCmd) version(cmd *cobra.Command, args []string) error {
	fmt.Println(versionStringWithOverrides(gitversion, gitcommit, builddate))
	return nil
}

// variables to be overridden by ldflags
var (
	gitversion string
	gitcommit  string
	builddate  string
)

// versionWithLDFlags takes in the overrides and returns a json compatible
// version string
func versionStringWithOverrides(gitversion string, gitcommit string, builddate string) string {
	v := defaultVersionInfo
	if gitversion != "" {
		v.GitVersion = gitversion
	}
	if gitcommit != "" {
		v.GitCommit = gitcommit
	}
	if builddate != "" {
		v.BuildDate = builddate
	}
	return fmt.Sprintf(`{"GitVersion":"%s","GitCommit":"%s","BuildDate":"%s","GoVersion":"%s","Platform":"%s","Compiler":"%s"}`, v.GitVersion, v.GitCommit, v.BuildDate, v.GoVersion, v.Platform, v.Compiler)
}
