package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/simontheleg/konf-go/prompt"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

type clientSetCreator = func(afero.Fs) (kubernetes.Interface, error)

type namespaceCmd struct {
	fs afero.Fs

	promptFunc       prompt.RunFunc
	selectNamespace  func(clientSetCreator, prompt.RunFunc, afero.Fs) (string, error)
	setNamespace     func(afero.Fs, string) error
	clientSetCreator clientSetCreator

	cmd *cobra.Command
}

func newNamespaceCmd() *namespaceCmd {

	fs := afero.NewOsFs()

	cc := &namespaceCmd{
		fs:               fs,
		promptFunc:       prompt.Terminal,
		selectNamespace:  selectNamespace,
		setNamespace:     setNamespace,
		clientSetCreator: newKubeClientSet,
	}

	cc.cmd = &cobra.Command{
		Use:     "namespace",
		Aliases: []string{"ns"},
		Short:   "Change namespace in current context",
		Long: `Set the namespace in the current context or start picker dialogue.
Can also be invoked via 'ns' alias

Examples:
-> 'ns' run namespace selection
-> 'ns <namespace-name' set to a specific namespace
`,
		RunE:              cc.namespace,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: cc.completeNamespace,
	}

	return cc
}

func (c *namespaceCmd) namespace(cmd *cobra.Command, args []string) error {
	var ns string
	var err error
	// TODO think about whether a setLastNamespace func should be implemented
	if len(args) == 0 {
		ns, err = c.selectNamespace(c.clientSetCreator, c.promptFunc, c.fs)
		if err != nil {
			return err
		}
	} else {
		ns = args[0]
	}

	err = c.setNamespace(c.fs, ns)
	if err != nil {
		return err
	}

	return nil
}

func (c *namespaceCmd) completeNamespace(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Note: we do not filter on our own using toComplete string as input.
	// This is so, that shells like zsh can use their completion matching to filter the list of namespaces
	// Essentially we are completely ignoring the toComplete argument and always return the full list
	// This is similar to how helm does it https://github.com/helm/helm/blob/35f6196167f6b90913f3a94237825856b858a215/cmd/helm/list.go#L227
	// Additionally, I think this makes it easier to integrate with another fuzzy parser like fzf later on

	// TODO the clientSetCreator logic for this and for selectNamespace could possibly be streamlined
	cs, err := c.clientSetCreator(c.fs)
	if err != nil {
		cobra.CompDebugln(err.Error(), true)
		return nil, cobra.ShellCompDirectiveError
	}

	nsl, err := cs.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	if err != nil {
		cobra.CompDebugln(err.Error(), true)
		return nil, cobra.ShellCompDirectiveError
	}

	nss := []string{}
	for _, ns := range nsl.Items {
		nss = append(nss, ns.Name)
	}

	return nss, cobra.ShellCompDirectiveNoFileComp
}

func selectNamespace(csc clientSetCreator, pf prompt.RunFunc, fs afero.Fs) (string, error) {
	cs, err := csc(fs)
	if err != nil {
		return "", err
	}

	nsl, err := cs.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	if err != nil {
		return "", err
	}

	nss := []string{}
	for _, ns := range nsl.Items {
		nss = append(nss, ns.Name)
	}

	// Wrapper is required as we need access to nss, but the methodSignature from promptUI
	// requires you to only pass an index not the whole func
	// This wrapper allows us to unit-test the SearchNamespace func
	var wrapSearchNamespace = func(input string, index int) bool {
		return searchNamespace(input, nss[index])
	}

	prompt := &promptui.Select{
		Label:        "Select namespace",
		Items:        nss,
		HideSelected: true,
		Stdout:       os.Stderr,
		Templates: &promptui.SelectTemplates{
			Active: fmt.Sprintf("%s {{ . | bold | cyan }}", promptui.IconSelect),
		},
		StartInSearchMode: true,
		Searcher:          wrapSearchNamespace,
		Size:              15,
	}

	selPos, err := pf(prompt)
	if err != nil {
		return "", err
	}

	if selPos >= len(nss) {
		return "", fmt.Errorf("invalid selection %d", selPos)
	}

	return nss[selPos], nil
}

func searchNamespace(searchTerm, curItem string) bool {
	return fuzzy.Match(searchTerm, curItem)
}

func newKubeClientSet(fs afero.Fs) (kubernetes.Interface, error) {
	kPath, err := kubeconfigEnv()
	if err != nil {
		return nil, err
	}

	b, err := afero.ReadFile(fs, kPath)
	if err != nil {
		return nil, err
	}

	conf, err := clientcmd.NewClientConfigFromBytes(b)
	if err != nil {
		return nil, err
	}

	cc, err := conf.ClientConfig()
	if err != nil {
		return nil, err
	}

	cs, err := kubernetes.NewForConfig(cc)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func setNamespace(fs afero.Fs, ns string) error {
	kPath, err := kubeconfigEnv()
	if err != nil {
		return err
	}

	b, err := afero.ReadFile(fs, kPath)
	if err != nil {
		return err
	}

	var conf k8s.Config
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return err
	}

	if len(conf.Contexts) == 0 {
		return fmt.Errorf("could not set namespace as contexts[] is empty in kubeconfig")
	}

	conf.Contexts[0].Context.Namespace = ns // this should be safe as konf import ensures we have only one context

	retconf, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	err = afero.WriteFile(fs, kPath, retconf, utils.KonfPerm)
	if err != nil {
		return err
	}

	return nil
}

func kubeconfigEnv() (string, error) {
	kPath := os.Getenv("KUBECONFIG")
	if kPath == "" {
		// it makes sense to return an error here, as depending funcs do not work without KUBECONFIG being set
		return "", fmt.Errorf("KUBECONFIG ist not set in your shell. Have you run konf set?")
	}
	return kPath, nil
}
