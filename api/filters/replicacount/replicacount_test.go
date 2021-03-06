package replicacount

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestFilter(t *testing.T) {
	var config = builtinconfig.MakeDefaultConfig()

	testCases := map[string]struct {
		input    string
		expected string
		filter   Filter
		fsslice  types.FsSlice
	}{
		"update field": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 5
`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 42
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "dep",
					Count: 42,
				},
			},
			fsslice: types.FsSlice{
				{
					Path: "spec/replicas",
				},
			},
		},
		"add field": {
			input: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
`,
			expected: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
    replicas: 42
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "cus",
					Count: 42,
				},
			},
			fsslice: types.FsSlice{
				{
					Path:               "spec/template/replicas",
					CreateIfNotPresent: true,
				},
			},
		},

		"add_field_null": {
			input: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
    replicas: null
`,
			expected: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
    replicas: 42
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "cus",
					Count: 42,
				},
			},
			fsslice: types.FsSlice{
				{
					Path:               "spec/template/replicas",
					CreateIfNotPresent: true,
				},
			},
		},
		"no update if CreateIfNotPresent is false": {
			input: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
`,
			expected: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  template:
    other: something
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "cus",
					Count: 42,
				},
			},
			fsslice: types.FsSlice{
				{
					Path: "spec/template/replicas",
				},
			},
		},
		"update multiple fields": {
			input: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  replicas: 5
  template:
    replicas: 5
`,
			expected: `
apiVersion: custom/v1
kind: Custom
metadata:
  name: cus
spec:
  replicas: 42
  template:
    replicas: 42
`,
			filter: Filter{
				Replica: types.Replica{
					Name:  "cus",
					Count: 42,
				},
			},
			fsslice: types.FsSlice{
				{
					Path: "spec/template/replicas",
				},
				{
					Path: "spec/replicas",
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			tc.filter.FsSlice = append(config.Replicas, tc.fsslice...)
			if !assert.Equal(t,
				strings.TrimSpace(tc.expected),
				strings.TrimSpace(
					filtertest_test.RunFilter(t, tc.input, tc.filter))) {
				t.FailNow()
			}
		})
	}
}
