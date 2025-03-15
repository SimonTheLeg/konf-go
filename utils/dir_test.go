package utils

import (
	"testing"

	"github.com/simontheleg/konf-go/config"
	"github.com/spf13/afero"
)

func TestEnsureDir(t *testing.T) {
	// since ensureDir is being run long before any storemanager is created,
	// we need to manually set a config here
	c := &config.Config{
		KonfDir: "./konf",
	}
	config.SetGlobalConfig(c)
	f := afero.NewMemMapFs()
	err := EnsureDir(f)
	if err != nil {
		t.Errorf("Unexpected error while running EnsureDir: %q", err)
	}

	r, err := f.Stat("./konf/active")
	if err != nil {
		t.Errorf("Could not run stat, please check tests: %v", err)
	}
	if r.IsDir() != true {
		t.Errorf("Expected %s to be a dir, but it is not %q", r.Name(), r)
	}

	r, err = f.Stat("./konf/store")
	if err != nil {
		t.Errorf("Could not run stat, please check tests: %v", err)
	}
	if r.IsDir() != true {
		t.Errorf("Expected %s to be a dir, but it is not %q", r.Name(), r)
	}
}
