package utils

import (
	"testing"

	"github.com/spf13/afero"
)

func TestEnsureDir(t *testing.T) {
	f := afero.NewMemMapFs()
	EnsureDir(f)

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
