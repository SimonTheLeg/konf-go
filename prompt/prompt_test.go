package prompt

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/simontheleg/konf-go/store"
)

func TestFuzzyFilterKonf(t *testing.T) {
	tt := map[string]struct {
		search string
		item   *store.TableOutput
		expRes bool
	}{
		"full match across all": {
			"a b c",
			&store.TableOutput{Context: "a", Cluster: "b", File: "c"},
			true,
		},
		"full match across all - fuzzy": {
			"abc",
			&store.TableOutput{Context: "a", Cluster: "b", File: "c"},
			true,
		},
		"partial match across fields": {
			"textclu",
			&store.TableOutput{Context: "context", Cluster: "cluster", File: "file"},
			true,
		},
		"no match": {
			"oranges",
			&store.TableOutput{Context: "apples", Cluster: "and", File: "bananas"},
			false,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			res := FuzzyFilterKonf(tc.search, tc.item)
			if res != tc.expRes {
				t.Errorf("Exp res to be %t got %t", tc.expRes, res)
			}
		})
	}
}

func TestPrepareTemplates(t *testing.T) {
	tt := map[string]struct {
		Values      store.TableOutput
		Trunc       int
		ExpInactive string
		ExpActive   string
		ExpLabel    string
	}{
		"values < trunc": {
			store.TableOutput{
				Context: "kind-eu",
				Cluster: "cluster-eu",
				File:    "kind-eu.cluster-eu.yaml",
			},
			25,
			"  kind-eu                   | cluster-eu                | kind-eu.cluster-eu.yaml   |",
			"▸ kind-eu                   | cluster-eu                | kind-eu.cluster-eu.yaml   |",
			"  Context                   | Cluster                   | File                      ",
		},
		"values == trunc": {
			store.TableOutput{
				Context: "0123456789",
				Cluster: "0123456789",
				File:    "xyz.yaml",
			},
			10,
			"  0123456789 | 0123456789 | xyz.yaml   |",
			"▸ 0123456789 | 0123456789 | xyz.yaml   |",
			"  Context    | Cluster    | File       ",
		},
		"values > trunc": {
			store.TableOutput{
				Context: "0123456789-andlotsmore",
				Cluster: "0123456789-andlotsmore",
				File:    "xyz.yaml",
			},
			10,
			"  0123456789 | 0123456789 | xyz.yaml   |",
			"▸ 0123456789 | 0123456789 | xyz.yaml   |",
			"  Context    | Cluster    | File       ",
		},
		"trunc is below minLength": {
			store.TableOutput{
				Context: "0123456789",
				Cluster: "0123456789",
				File:    "xyz.yaml",
			},
			5,
			"  0123456 | 0123456 | xyz.yam |",
			"▸ 0123456 | 0123456 | xyz.yam |",
			"  Context | Cluster | File    ",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			inactive, active, label := NewTableOutputTemplates(tc.Trunc)

			checkTemplate(t, inactive, tc.Values, tc.ExpInactive)
			checkTemplate(t, active, tc.Values, tc.ExpActive)
			checkTemplate(t, label, tc.Values, tc.ExpLabel)
		})
	}
}

func checkTemplate(t *testing.T, stpl string, val store.TableOutput, exp string) {

	tmpl, err := template.New("t").Funcs(NewStandardTemplateFuncs()).Parse(stpl)
	if err != nil {
		t.Fatalf("Could not create template for test '%v'. Please check test code", err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, val)
	if err != nil {
		t.Fatalf("Could not execute template for test '%v'. Please check test code", err)
	}

	res := buf.String()
	// remove any formatting as we do not care about that
	cyan := "\x1b[36m"
	bold := "\x1b[1m"
	normal := "\x1b[0m"
	res = strings.Replace(res, cyan, "", -1)
	res = strings.Replace(res, bold, "", -1)
	res = strings.Replace(res, normal, "", -1)
	if exp != res {
		t.Errorf("Exp res: '%s', got: '%s'", exp, res)
	}
}
