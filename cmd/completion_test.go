package cmd

import (
	"fmt"
	"testing"

	"github.com/simontheleg/konf-go/testhelper"
)

func TestCompletionCmd(t *testing.T) {

	tt := map[string]struct {
		args   []string
		ExpErr error
	}{
		"zsh arg": {
			[]string{"zsh"},
			nil,
		},
		"invalid bash arg": {
			[]string{"bash"},
			fmt.Errorf("konf currently does not support autocompletions for bash"),
		},
		"invalid arg": {
			[]string{"fish"},
			fmt.Errorf("konf currently does not support autocompletions for fish"),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			cc := newCompletionCmd()
			cmd := cc.cmd

			err := cmd.RunE(cmd, tc.args)

			if !testhelper.EqualError(err, tc.ExpErr) {
				t.Errorf("Want error '%s', got '%s'", tc.ExpErr, err)
			}
		})
	}
}
