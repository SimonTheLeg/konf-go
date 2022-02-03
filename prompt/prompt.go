package prompt

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

// RunFunc describes a generic function of a prompt. It returns the selected item.
// Its main purpose is to be easily mockable for unit-tests
type RunFunc func(*promptui.Select) (int, error)

// Terminal runs a given prompt in the terminal of the user and
// returns the selected items position
func Terminal(prompt *promptui.Select) (sel int, err error) {
	pos, _, err := prompt.Run()
	if err != nil {
		return -1, fmt.Errorf("prompt failed %v", err)
	}
	return pos, nil
}
