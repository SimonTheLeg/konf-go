package konf

import (
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/spf13/afero"
)

// IntegrationtestDir describes the directory to place files from IntegrationTests
const IntegrationtestDir = "/tmp/konfs"

var validCombos = []struct {
	context string
	cluster string
	id      KonfID
}{
	{"dev-eu", "dev-eu-1", "dev-eu_dev-eu-1"},
	{"con", "mygreathost.com-443", "con_mygreathost.com-443"},
	{"host.com-443/with/slashes", "danger", "host.com-443-with-slashes_danger"},
	{"host.com-443@something.nice", "danger", "host.com-443@something.nice_danger"},
	{"this:would:break:on:windows", "danger", "this-would-break-on-windows_danger"},
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
		Exp KonfID
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

func TestIDFromProcessID(t *testing.T) {
	in := 1234
	expOut := KonfID("1234")
	out := IDFromProcessID(in)
	if out != expOut {
		t.Errorf("Exp out to be %s, got %s", expOut, out)
	}
}

// this test simply checks if an ID is valid, by writing a file of that name to the os filesystem
// this test should be treated as an Integration test and run by CI on all OS supported by konf
func TestIDFileValidityIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestIDFileValidityIntegration integration test")
	}

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

	// should be ok to use the perm manually here, as we are inside an integration test
	// TODO later remove this part, as ID should have nothing to do with file storage
	err := f.MkdirAll(dir, 0700)
	if err != nil {
		t.Errorf("could not create dir for test %q", err)
	}

	for _, co := range validCombos {
		id := IDFromClusterAndContext(co.cluster, co.context)
		fpath := fmt.Sprintf("%s/%s.yaml", dir, id)

		// it should be fine to write empty
		err := afero.WriteFile(f, fpath, []byte{}, 0600)
		if err != nil {
			t.Errorf("Exp filename %q to work, but got error %q", fpath, err)
		}
	}
}
