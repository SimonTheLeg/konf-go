package cmd

import (
	"fmt"
	"testing"

	"github.com/simontheleg/konf-go/utils"
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
		"invalid arg": {
			[]string{"fish"},
			fmt.Errorf("konf currently does not support fish"),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			cs := newShellwrapperCmd()
			cs.cmd.SetArgs(tc.args)

			_, err := cs.cmd.ExecuteC()

			if !utils.EqualError(err, tc.ExpErr) {
				t.Errorf("Want error '%s', got '%s'", tc.ExpErr, err)
			}
		})
	}
}
