package prompt

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

// TODO maybe find a better name
type PromptFunc func(*promptui.Select) (int, error)

// terminal runs a given prompt in the terminal of the user and
// returns the selected items position
func Terminal(prompt *promptui.Select) (sel int, err error) {
	pos, _, err := prompt.Run()
	if err != nil {
		return -1, fmt.Errorf("prompt failed %v", err)
	}
	return pos, nil
}
