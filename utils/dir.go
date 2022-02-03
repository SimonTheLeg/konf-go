package utils

import (
	"io/fs"

	"github.com/simontheleg/konf-go/config"
	"github.com/spf13/afero"
)

// KonfPerm describes the file-permissions for konf files
const KonfPerm fs.FileMode = 0600 // based on the standard file-permissions for .kube/config

// KonfDirPerm describes the file-permissions for konf directories
const KonfDirPerm fs.FileMode = 0700 // needed so we can create folders inside

// IntegrationtestDir describes the directory to place files from IntegrationTests
const IntegrationtestDir = "/tmp/konfs"

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
