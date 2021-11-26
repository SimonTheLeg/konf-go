package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/simontheleg/konfig/utils"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const konfActiveList = "/.konf/active"

func initViper() {
	viper.Set("konfActiveList", konfActiveList)
}

func TestSelfClean(t *testing.T) {
	initViper()

	ppid := os.Getppid()

	tt := map[string]struct {
		Fs          afero.Fs
		ExpError    error
		ExpFiles    []string
		NotExpFiles []string
	}{
		"PID FS": {
			pidFS(),
			nil,
			[]string{konfActiveList + "/abc", konfActiveList + "/1234"},
			[]string{konfActiveList + "/" + fmt.Sprint(ppid)},
		},
		"PID file deleted by external source": {
			pidFileMissing(),
			nil,
			[]string{konfActiveList + "/abc", konfActiveList + "/1234"},
			[]string{},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {

			err := selfClean(tc.Fs)

			if !utils.EqualError(err, tc.ExpError) {
				t.Errorf("Want error '%s', got '%s'", tc.ExpError, err)
			}

			for _, s := range tc.ExpFiles {
				if _, err := tc.Fs.Stat(s); err != nil {
					t.Errorf("Exp file '%s' to exist, but it does not", s)
				}
			}

			for _, s := range tc.NotExpFiles {
				_, err := tc.Fs.Stat(s)

				if err == nil {
					t.Errorf("Exp file '%s' to be deleted, but it still exists", s)
				}

				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("An unexpected error has occured")
				}
			}
		})
	}
}

func pidFS() afero.Fs {
	ppid := os.Getppid()
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, konfActiveList+"/"+fmt.Sprint(ppid), []byte(singleClusterSingleContext), 0644)
	afero.WriteFile(fs, konfActiveList+"/abc", []byte("I am not event a kubeconfig, what am I doing here?"), 0644)
	afero.WriteFile(fs, konfActiveList+"/1234", []byte(singleClusterSingleContext), 0644)
	return fs
}

func pidFileMissing() afero.Fs {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, konfActiveList+"/abc", []byte("I am not event a kubeconfig, what am I doing here?"), 0644)
	afero.WriteFile(fs, konfActiveList+"/1234", []byte(singleClusterSingleContext), 0644)
	return fs
}
