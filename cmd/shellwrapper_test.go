package cmd

import (
	"fmt"
	"testing"

	"github.com/simontheleg/konf-go/testhelper"
)

func TestShellWrapperCmd(t *testing.T) {

	tt := map[string]struct {
		args   []string
		ExpErr error
	}{
		"zsh arg": {
			[]string{"zsh"},
			nil,
		},
		"bash arg": {
			[]string{"bash"},
			nil,
		},
		"fish arg": {
			[]string{"fish"},
			nil,
		},
		"invalid arg": {
			[]string{"invalid"},
			fmt.Errorf("konf currently does not support invalid"),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			cs := newShellwrapperCmd()
			cmd := cs.cmd

			err := cmd.RunE(cmd, tc.args)

			if !testhelper.EqualError(err, tc.ExpErr) {
				t.Errorf("Want error '%s', got '%s'", tc.ExpErr, err)
			}
		})
	}
}
