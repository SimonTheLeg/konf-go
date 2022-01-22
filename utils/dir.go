package utils

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// EnsureDir makes sure that konf store and active dirs exist
func EnsureDir(f afero.Fs) error {

	err := f.MkdirAll(viper.GetString("storeDir")+"/", 0600)
	if err != nil {
		return err
	}

	err = f.MkdirAll(viper.GetString("activeDir")+"/", 0600)
	if err != nil {
		return err
	}

	return nil
}
