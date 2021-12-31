package utils

import (
	"io/fs"
	"testing"
	"time"
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
	cl, co, exp := "abc", "xyz", "abc_xyz"
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
