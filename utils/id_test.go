package utils

import (
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/spf13/afero"
)

var validCombos = []struct {
	context string
	cluster string
	id      string
}{
	{"dev-eu", "dev-eu-1", "dev-eu_dev-eu-1"},
	{"con", "mygreathost.com-443", "con_mygreathost.com-443"},
	{"host.com-443/with/slashes", "danger", "host.com-443-with-slashes_danger"},
	{"host.com-443@something.nice", "danger", "host.com-443@something.nice_danger"},
	{"this:would:break:on:windows", "danger", "this-would-break-on-windows_danger"},
}

func TestPathForID(t *testing.T) {
	for _, co := range validCombos {
		resStore := StorePathForID(co.id)
		expStore := fmt.Sprintf("./konf/store/%s.yaml", co.id)
		if resStore != expStore {
			t.Errorf("Exp StorePath %q, got %q", expStore, resStore)
		}

		resActive := ActivePathForID(co.id)
		expActive := fmt.Sprintf("./konf/active/%s.yaml", co.id)
		if resActive != expActive {
			t.Errorf("Exp ActivePath %q, got %q", expActive, resActive)
		}
	}

}

func TestIDFromClusterAndContext(t *testing.T) {
	for _, co := range validCombos {
		res := IDFromClusterAndContext(co.cluster, co.context)
		if res != co.id {
			t.Errorf("Exp ID %q, got %q", co.id, res)
		}
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

	for _, co := range validCombos {
		id := IDFromClusterAndContext(co.cluster, co.context)
		fpath := fmt.Sprintf("%s/%s.yaml", dir, id)

		err := afero.WriteFile(f, fpath, []byte(sm.SingleClusterSingleContextEU()), KonfPerm)
		if err != nil {
			t.Errorf("Exp filename %q to work, but got error %q", fpath, err)
		}
	}
}
