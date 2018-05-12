/*
Copyright 2018 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubectl

import (
	"os"
	"testing"

	"github.com/AmitKumarDas/elitmus/pkg/util"
)

func TestGetKubectlPath(t *testing.T) {
	if envVal := util.KubectlPathENV(); len(envVal) != 0 {
		os.Unsetenv(string(util.KubectlPathENVK))
		defer func() { os.Setenv(string(util.KubectlPathENVK), envVal) }()
	}

	tests := map[string]struct {
		kubectlPathVal string
	}{
		"get kubectl path: positive test case - with env set - 1": {
			kubectlPathVal: "Hi",
		},
		"get kubectl path: positive test case - with env set - 2": {
			kubectlPathVal: "There",
		},
		"get kubectl path: poisitve test case - env not set": {
			// no value will be set in env
			kubectlPathVal: "",
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// set a test value in env before testing
			os.Setenv(string(util.KubectlPathENVK), mock.kubectlPathVal)
			// function under test
			p := GetKubectlPath()

			if len(mock.kubectlPathVal) != 0 && mock.kubectlPathVal != p {
				t.Fatalf("failed to get kubectl path: expected '%s': actual '%s'", mock.kubectlPathVal, p)
			}

			if len(mock.kubectlPathVal) == 0 && KubectlPath != p {
				t.Fatalf("failed to get kubectl path: expected '%s': actual '%s'", KubectlPath, p)
			}
		})
	}
}

func TestKubeCtlArgs(t *testing.T) {
	tests := map[string]struct {
		args      []string
		namespace string
		context   string
		labels    string
		expected  []string
		isEmpty   bool
	}{
		"kubectl args - positive test case": {
			args:      []string{"kubectl", "get", "po", "my-pod"},
			namespace: "litmus",
			context:   "",
			labels:    "name=my-pod",
			expected:  []string{"kubectl", "get", "po", "my-pod", "--namespace=litmus", "--selector=name=my-pod"},
			isEmpty:   false,
		},
		"kubectl args - negative test case - empty": {
			args:      []string{},
			namespace: "",
			context:   "",
			labels:    "",
			expected:  []string{},
			isEmpty:   true,
		},
		"kubectl args - negative test case - nil": {
			args:      nil,
			namespace: "",
			context:   "",
			labels:    "",
			expected:  nil,
			isEmpty:   true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ops := kubectlArgs(mock.args, mock.namespace, mock.context, mock.labels)

			if len(ops) == 0 && !mock.isEmpty {
				t.Fatalf("failed to execute kubectl args: expected 'non nil output': actual 'nil output'")
			}

			if len(ops) != len(mock.expected) {
				t.Fatalf("failed to execute kubectl args: expected output '%s': actual output '%s'", mock.expected, ops)
			}
		})
	}
}
