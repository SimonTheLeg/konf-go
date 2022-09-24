package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"
)

func TestVersionStringWithOverrides(t *testing.T) {
	tt := map[string]struct {
		gitversion string
		gitcommit  string
		builddate  string
		exp        string
	}{
		"no overrides": {
			exp: fmt.Sprintf(`{"GitVersion":"dev","GitCommit":"dev","BuildDate":"1970-01-01T00:00:00Z","GoVersion":"%s","Platform":"%s","Compiler":"%s"}`, runtime.Version(), fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH), runtime.Compiler),
		},
		"gitversion override": {
			gitversion: "override",
			exp:        fmt.Sprintf(`{"GitVersion":"override","GitCommit":"dev","BuildDate":"1970-01-01T00:00:00Z","GoVersion":"%s","Platform":"%s","Compiler":"%s"}`, runtime.Version(), fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH), runtime.Compiler),
		},
		"gitcommit override": {
			gitcommit: "override",
			exp:       fmt.Sprintf(`{"GitVersion":"dev","GitCommit":"override","BuildDate":"1970-01-01T00:00:00Z","GoVersion":"%s","Platform":"%s","Compiler":"%s"}`, runtime.Version(), fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH), runtime.Compiler),
		},
		"builddate override": {
			builddate: "override",
			exp:       fmt.Sprintf(`{"GitVersion":"dev","GitCommit":"dev","BuildDate":"override","GoVersion":"%s","Platform":"%s","Compiler":"%s"}`, runtime.Version(), fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH), runtime.Compiler),
		},
		"all values override": {
			gitversion: "override",
			gitcommit:  "override",
			builddate:  "override",
			exp:        fmt.Sprintf(`{"GitVersion":"override","GitCommit":"override","BuildDate":"override","GoVersion":"%s","Platform":"%s","Compiler":"%s"}`, runtime.Version(), fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH), runtime.Compiler),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			res := versionStringWithOverrides(tc.gitversion, tc.gitcommit, tc.builddate)

			if res != tc.exp {
				t.Errorf("Exp res to be '%s', got '%s'", tc.exp, res)
			}

			// check if the result is a valid json
			js := json.RawMessage{}
			if err := json.Unmarshal([]byte(res), &js); err != nil {
				t.Errorf("Exp to unmarshal version string without error, but got '%v'", err)
			}
		})
	}

}
