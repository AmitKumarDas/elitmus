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

package fetch

import (
	"fmt"
	"testing"

	"github.com/AmitKumarDas/litmus/pkg/kubectl"
	"github.com/AmitKumarDas/litmus/pkg/meta"
)

func TestInstanceCreation(t *testing.T) {
	tests := map[string]struct {
		installFile meta.InstallFile
		isErr       bool
	}{
		"new instance of KubeResourceFetch - negative test case - non existent install file": {
			installFile: meta.InstallFile("/home/non-existent.txt"),
			isErr:       true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k, err := NewKubeResourceFetch(mock.installFile)

			if err != nil && !mock.isErr {
				t.Fatalf("failed to verify instance creation: expected 'no error': actual '%#v'", err)
			}

			if k == nil && !mock.isErr {
				t.Fatalf("failed to verify instance creation: expected 'non nil instance': actual 'nil instance'")
			}
		})
	}
}

func TestFetch(t *testing.T) {
	tests := map[string]struct {
		alias    string
		property Property
		isErr    bool
	}{
		"fetch - negative test case - unsupported property": {
			alias: "service",
			// this property is not supported; hence negative test case
			property: Property("mock"),
			isErr:    true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := &KubeResourceFetch{}

			d, err := k.Fetch(mock.alias, mock.property)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to verify fetch: expected 'no error': actual '%#v'", err)
			}

			if err != nil && mock.isErr {
				expectedErr := fmt.Sprintf("property '%s' is not supported by kubernetes resource", mock.property)
				if err.Error() != expectedErr {
					t.Fatalf("failed to verify fetch: expected error '%s': actual error '%s'", expectedErr, err.Error())
				}
			}

			if len(d) == 0 && !mock.isErr {
				t.Fatalf("failed to verify fetch: expected 'non zero len output': actual 'zero len output'")
			}
		})
	}
}

func TestFetchServiceIP(t *testing.T) {
	tests := map[string]struct {
		installation   *meta.Installation
		alias          string
		kubectlFactory kubectl.KubeFactory
		isErr          bool
	}{
		"fetch service ip - positive test case": {
			installation: &meta.Installation{
				Components: []meta.Component{
					meta.Component{
						Name:      "MyService",
						Namespace: "TestNS",
						Kind:      "service",
						Alias:     "coolservice",
					},
				},
			},
			alias:          "coolservice",
			kubectlFactory: &kubectl.MockKubeFactory{},
			isErr:          false,
		},
		"fetch service ip - negative test case - no matching alias": {
			installation: &meta.Installation{
				Components: []meta.Component{
					meta.Component{
						Name:      "MyService",
						Namespace: "TestNS",
						Kind:      "service",
						Alias:     "coolservice",
					},
				},
			},
			// alias does not match with installation's alias; hence negative test case
			alias:          "coolsvc",
			kubectlFactory: &kubectl.MockKubeFactory{},
			isErr:          true,
		},
		"fetch service ip - negative test case - no output": {
			installation: &meta.Installation{
				Components: []meta.Component{
					meta.Component{
						Name:      "MyService",
						Namespace: "TestNS",
						Kind:      "service",
						Alias:     "coolservice",
					},
				},
			},
			alias: "coolservice",
			// mock that does not return any output; hence negative test case
			kubectlFactory: &kubectl.MockKubeFactoryNoOutput{},
			isErr:          true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := &KubeResourceFetch{
				installation:   mock.installation,
				kubectlFactory: mock.kubectlFactory,
			}

			d, err := k.fetchServiceIP(mock.alias)

			if err != nil && !mock.isErr {
				t.Fatalf("failed to verify fetch service ip: expected 'no error': actual '%#v'", err)
			}

			if len(d) == 0 && !mock.isErr {
				t.Fatalf("failed to verify fetch service ip: expected 'non zero len output': actual 'zero len output'")
			}
		})
	}
}
