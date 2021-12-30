package utils

import "testing"

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
