package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/manifoldco/promptui"
	"github.com/simontheleg/konf-go/prompt"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   `set`,
	Short: "Set kubeconfig to use in current shell",
	Args:  cobra.MaximumNArgs(1),
	Long: `Sets kubeconfig to use or start picker dialogue.

Examples:
	-> 'set' run konf selection
	-> 'set <konfig id>' set a specific konf
	-> 'set -' set to last used konf
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var id string
		f := afero.NewOsFs()
		var err error

		if len(args) == 0 {
			id, err = selectContext(f, prompt.Terminal)
			if err != nil {
				return err
			}
		} else if args[0] == "-" {
			id, err = selectLastKonf(f)
			if err != nil {
				return err
			}
		} else {
			id = args[0]
		}

		context, err := setContext(id, f)
		if err != nil {
			return err
		}
		err = saveLatestKonf(f, id)
		if err != nil {
			return fmt.Errorf("could not save latest konf. As a result 'konf set -' might not work: %q ", err)
		}

		log.Printf("Setting context to %q\n", id)
		// By printing out to stdout, we pass the value to our zsh hook, which then sets $KUBECONFIG to it
		fmt.Println(context)

		return nil
	},
}

type promptFunc func(*promptui.Select) (int, error)

func selectContext(f afero.Fs, pf promptFunc) (string, error) {
	k, err := fetchKonfs(f)
	if err != nil {
		return "", err
	}
	p := createPrompt(k)
	selPos, err := pf(p)
	if err != nil {
		return "", err
	}

	if selPos >= len(k) {
		return "", fmt.Errorf("invalid selection %d", selPos)
	}
	sel := k[selPos]

	return utils.IDFromClusterAndContext(sel.Cluster, sel.Context), nil
}

func selectLastKonf(f afero.Fs) (string, error) {
	b, err := afero.ReadFile(f, viper.GetString("latestKonfFile"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("could not select latest konf, because no konf was yet set")
		} else {
			return "", err
		}
	}
	return string(b), nil
}

func setContext(id string, f afero.Fs) (string, error) {
	konf, err := afero.ReadFile(f, utils.StorePathForID(id))
	if err != nil {
		return "", err
	}

	ppid := os.Getppid()
	activeKonf := utils.ActivePathForID(fmt.Sprint(ppid))
	err = afero.WriteFile(f, activeKonf, konf, utils.KonfPerm)
	if err != nil {
		return "", err
	}

	return activeKonf, nil

}

func saveLatestKonf(f afero.Fs, id string) error {
	return afero.WriteFile(f, viper.GetString("latestKonfFile"), []byte(id), utils.KonfPerm)
}

// KubeConfigOverload describes a state in which a kubeconfig has multiple Contexts or Clusters
// This can be undesirable for konf when such a kubeconfig is in its store
type KubeConfigOverload struct {
	path string
}

func (k *KubeConfigOverload) Error() string {
	return fmt.Sprintf("Impure Store: The kubeconfig %q contains multiple contexts and/or clusters. Please only use 'konf import' for populating the store\n", k.path)
}

// EmptyStore describes a state in which no kubeconfig is inside the store
// It makes sense to have this in a separate case as it does not matter for some operations (e.g. importing) but detrimental for others (e.g. running the selection prompt)
type EmptyStore struct{}

func (k *EmptyStore) Error() string {
	return fmt.Sprintf("The konf store at %q is empty. Please run 'konf import' to populate it", viper.GetString("storeDir"))
}

// fetchKonfs returns a list of all konfs currently in konfDir/store. Additionally it returns metadata on these konfs for easier usage of the information
func fetchKonfs(f afero.Fs) ([]tableOutput, error) {
	konfs, err := afero.ReadDir(f, viper.GetString("storeDir"))
	if err != nil {
		return nil, err
	}

	if len(konfs) == 0 {
		return nil, &EmptyStore{}
	}

	out := []tableOutput{}
	for _, konf := range konfs {
		id := utils.IDFromFileInfo(konf)
		path := utils.StorePathForID(id)
		file, err := f.Open(path)
		if err != nil {
			return nil, err
		}
		val, err := afero.ReadAll(file)
		if err != nil {
			return nil, err
		}
		kubeconf := &k8s.Config{}
		err = yaml.Unmarshal(val, kubeconf)
		if err != nil {
			log.Printf("file %q does not contain a valid kubeconfig. Skipping for evaluation", path)
			continue
		}

		if len(kubeconf.Contexts) > 1 || len(kubeconf.Clusters) > 1 {
			// This directly returns, as an impure store is a danger for other usage down the road
			return nil, &KubeConfigOverload{path}
		}

		t := tableOutput{}
		t.Context = kubeconf.Contexts[0].Name
		t.Cluster = kubeconf.Clusters[0].Name
		t.File = path
		out = append(out, t)
	}
	return out, nil
}

func createPrompt(options []tableOutput) *promptui.Select {
	// TODO use ssh/terminal to get the terminalsize and set trunc accordingly https://stackoverflow.com/questions/16569433/get-terminal-size-in-go
	trunc := 25
	promptInactive, promptActive, label := prepareTable(trunc)

	prompt := promptui.Select{
		Label: label,
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   promptActive,
			Inactive: promptInactive,
			FuncMap:  newTemplateFuncMap(),
		},
		HideSelected: true,
		Stdout:       os.Stderr,
	}
	return &prompt
}

// TODO only inject the funcs I am actually using
func newTemplateFuncMap() template.FuncMap {
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

// tableOutput describes a formatting of kubekonf information, that is being used to present the user a nice table selection
type tableOutput struct {
	// Since we have no other use for structured information, we can safely leave this in set.go for now
	Context string
	Cluster string
	File    string
}

// prepareTable takes in the max length of each column and returns table rows for active, inactive and header
func prepareTable(maxColumnLen int) (inactive, active, label string) {
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

func init() {
	setCmd.SetOut(os.Stderr)
	rootCmd.AddCommand(setCmd)
}
