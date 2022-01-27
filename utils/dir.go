package utils

import (
	"io/fs"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const KonfPerm fs.FileMode = 0600 // based on the standard file-permissions for .kube/config
const KonfDirPerm fs.FileMode = 0700 // needed so we can create folders inside

// EnsureDir makes sure that konf store and active dirs exist
func EnsureDir(f afero.Fs) error {

	err := f.MkdirAll(viper.GetString("storeDir")+"/", KonfDirPerm)
	if err != nil {
		return err
	}

	err = f.MkdirAll(viper.GetString("activeDir")+"/", KonfDirPerm)
	if err != nil {
		return err
	}

	return nil
}
