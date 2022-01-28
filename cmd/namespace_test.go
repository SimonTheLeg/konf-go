package cmd

import (
	"fmt"
	"testing"

	"github.com/manifoldco/promptui"
	"github.com/simontheleg/konf-go/prompt"
	"github.com/simontheleg/konf-go/testhelper"
	"github.com/spf13/afero"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
)

func TestNamespace(t *testing.T) {

	selectNamespaceCalled := false
	setNamespaceCalled := false
	var mockSelectNamespace = func(clientSetCreator, prompt.PromptFunc, afero.Fs) (string, error) {
		selectNamespaceCalled = true
		return "", nil
	}
	var mockSetNamespace = func(afero.Fs, string) error { setNamespaceCalled = true; return nil }

	nscmd := newNamespaceCmd()
	nscmd.selectNamespace = mockSelectNamespace
	nscmd.setNamespace = mockSetNamespace

	type ExpCalls struct {
		SelectNamespace bool
		SetNamespace    bool
	}
	tt := map[string]struct {
		Args   []string
		ExpErr error
		ExpCalls
	}{
		"1 arg": {
			[]string{"ns1"},
			nil,
			ExpCalls{SelectNamespace: false, SetNamespace: true},
		},
		"0 args": {
			[]string{},
			nil,
			ExpCalls{SelectNamespace: true, SetNamespace: true},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			selectNamespaceCalled = false
			setNamespaceCalled = false
			cmd := nscmd.cmd

			err := cmd.RunE(cmd, tc.Args)
			if !testhelper.EqualError(tc.ExpErr, err) {
				t.Errorf("Exp error %q, got %q", tc.ExpErr, err)
			}

			if tc.ExpCalls.SelectNamespace != selectNamespaceCalled {
				t.Errorf("Exp SelectNamespaceCalled to be %t, but got %t", tc.ExpCalls.SelectNamespace, selectNamespaceCalled)
			}

			if tc.ExpCalls.SetNamespace != setNamespaceCalled {
				t.Errorf("Exp SetNamespaceCalled to be %t, but got %t", tc.ExpCalls.SetNamespace, setNamespaceCalled)
			}

		})
	}
}

func TestSearchNamespace(t *testing.T) {
	tt := map[string]struct {
		search string
		item   string
		expRes bool
	}{
		"full-match": {
			"kube-system",
			"kube-system",
			true,
		},
		"partial-match-front": {
			"kube",
			"kube-system",
			true,
		},
		"partial-match-middle": {
			"e-sys",
			"kube-system",
			true,
		},
		"partial-match-end": {
			"stem",
			"kube-system",
			true,
		},
		"no-match": {
			"apples and oranges",
			"kube-system",
			false,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			res := searchNamespace(tc.search, tc.item)
			if res != tc.expRes {
				t.Errorf("Exp res to be %t got %t", tc.expRes, res)
			}
		})
	}
}

func TestNewKubeClientSet(t *testing.T) {
	fm := testhelper.FilesystemManager{}

	tt := map[string]struct {
		kubeenv string
		Fs      afero.Fs
		ExpErr  bool
	}{
		"no $KUBECONFIG set": {
			"",
			nil,
			true,
		},
		"valid kubeconfig": {
			"./konf/active/dev-eu_dev-eu-1.yaml",
			testhelper.FSWithFiles(fm.ActiveDir, fm.SingleClusterSingleContextEU),
			false,
		},
		"invalid kubeconfig": {
			"./konf/active/no-konf.yaml",
			testhelper.FSWithFiles(fm.ActiveDir, fm.InvalidYaml),
			true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			t.Setenv("KUBECONFIG", tc.kubeenv)

			_, err := newKubeClientSet(tc.Fs)

			if err != nil && tc.ExpErr == false {
				t.Errorf("Exp no error, but got: %v", err)
			}
		})
	}
}

func TestSelectNamespace(t *testing.T) {

	// keep these in alphabetical order for tests to work!
	nss := []runtime.Object{
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "first",
			},
		},
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kube-system",
			},
		},
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "zebra",
			},
		},
	}

	var mockClientSetCreator = func(afero.Fs) (kubernetes.Interface, error) {
		return fake.NewSimpleClientset(nss...), nil
	}

	var mockSelect = func(sel int) prompt.PromptFunc {
		return func(*promptui.Select) (int, error) {
			return sel, nil
		}
	}

	tt := map[string]struct {
		csc    clientSetCreator
		sel    func(*promptui.Select) (int, error)
		expNS  string
		expErr error
	}{
		"valid selection": {
			mockClientSetCreator,
			mockSelect(1),
			"kube-system",
			nil,
		},
		"invalid selection": {
			mockClientSetCreator,
			mockSelect(3),
			"",
			fmt.Errorf("invalid selection 3"),
		},
		"error prompt": {
			mockClientSetCreator,
			func(s *promptui.Select) (int, error) { return 0, fmt.Errorf("big bad error") },
			"",
			fmt.Errorf("big bad error"),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			res, err := selectNamespace(tc.csc, tc.sel, nil)

			if !testhelper.EqualError(err, tc.expErr) {
				t.Errorf("Exp err %q, got %q", tc.expErr, err)
			}

			if res != tc.expNS {
				t.Errorf("Exp namespace to be %q, got %q", tc.expNS, res)
			}
		})
	}

}

func TestSetNamespace(t *testing.T) {
	fm := testhelper.FilesystemManager{}

	tt := map[string]struct {
		kubeenv string
		Fs      afero.Fs
		ns      string
		ExpErr  bool
	}{
		"no $KUBECONFIG set": {
			"",
			nil,
			"",
			true,
		},
		"valid kubeconfig": {
			"./konf/active/dev-eu_dev-eu-1.yaml",
			testhelper.FSWithFiles(fm.ActiveDir, fm.SingleClusterSingleContextEU),
			"kube-system",
			false,
		},
		"invalid kubeconfig": {
			"./konf/active/no-konf.yaml",
			testhelper.FSWithFiles(fm.ActiveDir, fm.InvalidYaml),
			"kube-system",
			true,
		},
		"valid kubeconfig, but missing context[]": {
			"./konf/active/no-context.yaml",
			testhelper.FSWithFiles(fm.ActiveDir, fm.KonfWithoutContext),
			"kube-system",
			true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			t.Setenv("KUBECONFIG", tc.kubeenv)

			err := setNamespace(tc.Fs, tc.ns)

			if err != nil && tc.ExpErr == false {
				t.Errorf("Exp no error, but got: %v", err)
			}

			if tc.ExpErr == false {
				b, err := afero.ReadFile(tc.Fs, tc.kubeenv)
				if err != nil {
					t.Errorf("failed to read file %q", err)
				}

				var kconf k8s.Config
				err = yaml.Unmarshal(b, &kconf)
				if err != nil {
					t.Errorf("failed to unmarshal %q", err)
				}

				resNs := kconf.Contexts[0].Context.Namespace
				if resNs != tc.ns {
					t.Errorf("exp ns to be %q, but is %q", tc.ns, resNs)
				}
			}
		})
	}
}
