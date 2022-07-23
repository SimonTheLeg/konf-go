package prompt

import (
	"fmt"
	"strings"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/simontheleg/konf-go/store"
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

// FuzzyFilterKonf allows fuzzy searching of a list of konf metadata in the form of store.TableOutput
func FuzzyFilterKonf(searchTerm string, curItem *store.TableOutput) bool {
	// since there is no weight on any of the table entries, we can just combine them to one string
	// and run the contains on it, which automatically is going to match any of the three values
	r := fmt.Sprintf("%s %s %s", curItem.Context, curItem.Cluster, curItem.File)
	return fuzzy.Match(searchTerm, r)
}

// NewStandardTemplateFuncs creates a template.FuncMap with all the standard
// template funcs (e.g. trunc, cyan, ...) and returns them
// TODO only inject the funcs I am actually using
func NewStandardTemplateFuncs() template.FuncMap {
	ret := sprig.TxtFuncMap()
	ret["black"] = promptui.Styler(promptui.FGBlack)
	ret["red"] = promptui.Styler(promptui.FGRed)
	ret["green"] = promptui.Styler(promptui.FGGreen)
	ret["yellow"] = promptui.Styler(promptui.FGYellow)
	ret["blue"] = promptui.Styler(promptui.FGBlue)
	ret["magenta"] = promptui.Styler(promptui.FGMagenta)
	ret["cyan"] = promptui.Styler(promptui.FGCyan)
	ret["white"] = promptui.Styler(promptui.FGWhite)
	ret["bgBlack"] = promptui.Styler(promptui.BGBlack)
	ret["bgRed"] = promptui.Styler(promptui.BGRed)
	ret["bgGreen"] = promptui.Styler(promptui.BGGreen)
	ret["bgYellow"] = promptui.Styler(promptui.BGYellow)
	ret["bgBlue"] = promptui.Styler(promptui.BGBlue)
	ret["bgMagenta"] = promptui.Styler(promptui.BGMagenta)
	ret["bgCyan"] = promptui.Styler(promptui.BGCyan)
	ret["bgWhite"] = promptui.Styler(promptui.BGWhite)
	ret["bold"] = promptui.Styler(promptui.FGBold)
	ret["faint"] = promptui.Styler(promptui.FGFaint)
	ret["italic"] = promptui.Styler(promptui.FGItalic)
	ret["underline"] = promptui.Styler(promptui.FGUnderline)
	return ret
}

// NewTableOutputTemplates returns templating strings for creating a nicely
// formatted table out of an store.TableOutput. Maximum length per comment can
// be configured.
func NewTableOutputTemplates(maxColumnLen int) (inactive, active, label string) {
	// minColumnLen is determined by the length of the largest word in the label line
	minColumnLen := 7
	if maxColumnLen < minColumnLen {
		maxColumnLen = minColumnLen
	}
	// TODO figure out if we can do abbreviation using '...' somehow
	inactive = fmt.Sprintf(`  {{ repeat %[1]d " " | print .Context | trunc %[1]d | %[2]s }} | {{ repeat %[1]d " " | print .Cluster | trunc %[1]d | %[2]s }} | {{ repeat %[1]d  " " | print .File | trunc %[1]d | %[2]s }} |`, maxColumnLen, "")
	active = fmt.Sprintf(`▸ {{ repeat %[1]d " " | print .Context | trunc %[1]d | %[2]s }} | {{ repeat %[1]d " " | print .Cluster | trunc %[1]d | %[2]s }} | {{ repeat %[1]d  " " | print .File | trunc %[1]d | %[2]s }} |`, maxColumnLen, "bold | cyan")
	label = fmt.Sprint("  Context" + strings.Repeat(" ", maxColumnLen-7) + " | " + "Cluster" + strings.Repeat(" ", maxColumnLen-7) + " | " + "File" + strings.Repeat(" ", maxColumnLen-4) + " ") // repeat = trunc - length of the word before it
	return inactive, active, label
}
