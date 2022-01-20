// +build windows

package utils

import (
	"testing"
	"strings"
)

func TestIDFromClusterAndContext(t *testing.T) {
	cl, co, exp := "cluster", "context", "context_cluster"
	res := IDFromClusterAndContext(cl, co)
	if res != exp {
		t.Errorf("Exp ID %q, got %q", exp, res)
	}

	// TODO verify on a windows machine if this is working
	// test for windows with a name longer than 120 characters
	cl, co, exp = strings.Repeat("a", 120), "someothercontext", strings.Repeat("a", 120)
	res = IDFromClusterAndContext(cl, co)
	if res != exp {
		t.Errorf("Exp ID %q, got %q", exp, res)
	}
}
