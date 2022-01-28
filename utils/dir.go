package utils

import (
	"io/fs"

	"github.com/simontheleg/konf-go/config"
	"github.com/spf13/afero"
)

const KonfPerm fs.FileMode = 0600    // based on the standard file-permissions for .kube/config
const KonfDirPerm fs.FileMode = 0700 // needed so we can create folders inside

// EnsureDir makes sure that konf store and active dirs exist
func EnsureDir(f afero.Fs) error {

	err := f.MkdirAll(config.StoreDir()+"/", KonfDirPerm)
	if err != nil {
		return err
	}

	err = f.MkdirAll(config.ActiveDir()+"/", KonfDirPerm)
	if err != nil {
		return err
	}

	return nil
}
