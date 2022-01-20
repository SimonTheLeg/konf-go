// +build !windows

package utils

import (
	"testing"
)

func TestIDFromClusterAndContext(t *testing.T) {
	cl, co, exp := "cluster", "context", "context_cluster"
	res := IDFromClusterAndContext(cl, co)
	if res != exp {
		t.Errorf("Exp ID %q, got %q", exp, res)
	}
}