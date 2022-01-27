package utils

import (
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/spf13/afero"
)

// As we currently do not have any requirements for what an ID looks like, these unit tests are more for coverage sake

var allowedIDs = []string{
	"i-have-dashes",
	"i/have/slashes",
	"i.have.dots",
}

func TestPathForID(t *testing.T) {
	InitTestViper()

	expStorePaths := []string{
		"./konf/store/i-have-dashes.yaml",
		"./konf/store/i/have/slashes.yaml",
		"./konf/store/i.have.dots.yaml",
	}

	for i := range allowedIDs {
		res := StorePathForID(allowedIDs[i])
		exp := expStorePaths[i]
		if res != expStorePaths[i] {
			t.Errorf("Exp StorePath %q, got %q", exp, res)
		}
	}

	expActivePaths := []string{
		"./konf/active/i-have-dashes.yaml",
		"./konf/active/i/have/slashes.yaml",
		"./konf/active/i.have.dots.yaml",
	}

	for i := range allowedIDs {
		res := ActivePathForID(allowedIDs[i])
		exp := expActivePaths[i]
		if res != expActivePaths[i] {
			t.Errorf("Exp StorePath %q, got %q", exp, res)
		}
	}
}

func TestIDFromClusterAndContext(t *testing.T) {
	cl, co, exp := "cluster", "context", "context_cluster"
	res := IDFromClusterAndContext(cl, co)
	if res != exp {
		t.Errorf("Exp ID %q, got %q", exp, res)
	}
}

type mockFileInfo struct{ name string }

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() fs.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }

func TestIDFromFileInfo(t *testing.T) {

	tt := map[string]struct {
		In  fs.FileInfo
		Exp string
	}{
		"yaml extension": {
			&mockFileInfo{"mygreatid.yaml"},
			"mygreatid",
		},
		"no extension": {
			&mockFileInfo{"noextension"},
			"noextension",
		},
		"some other extension": {
			&mockFileInfo{"mygreatid.json"},
			"mygreatid",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			res := IDFromFileInfo(tc.In)
			if res != tc.Exp {
				t.Errorf("Expected ID %q, got %q", tc.Exp, res)
			}
		})
	}
}

// this test simply checks if an ID is valid, by writing a file of that name to the os filesystem
// this test should be treated as an Integration test and run by CI on all OS supported by konf
func TestIDFileValidityIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestIDFileValidityIntegration integration test")
	}

	sm := SampleKonfManager{}
	f := afero.NewOsFs()
	dir := IntegrationtestDir + "/TestIDFileValidityIntegration"

	t.Cleanup(
		func() {
			err := f.RemoveAll(IntegrationtestDir)
			if err != nil {
				t.Errorf("Cleanup failed %q", err)
			}
		},
	)

	err := f.MkdirAll(dir, KonfDirPerm)
	if err != nil {
		t.Errorf("could not create dir for test %q", err)
	}

	validCombos := []struct {
		context string
		cluster string
	}{
		{"dev-eu", "dev-eu-1"},
		{"con", "mygreathost.com:443"},
	}

	for _, co := range validCombos {
		id := IDFromClusterAndContext(co.cluster, co.context)
		fpath := fmt.Sprintf("%s/%s.yaml", dir, id)

		err := afero.WriteFile(f, fpath, []byte(sm.SingleClusterSingleContextEU()), KonfPerm)
		if err != nil {
			t.Errorf("Exp filename %q to work, but got error %q", fpath, err)
		}
	}
}
