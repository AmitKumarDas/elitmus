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

type mockKubectl struct{}

func (k *mockKubectl) Run(args []string) (output string, err error) {
	output = "mocked"
	return
}

func (k *mockKubectl) StdinRun(args []string, stdin []byte) (output string, err error) {
	output = "mocked"
	return
}

type mockKubectlNoOutput struct{}

func (k *mockKubectlNoOutput) Run(args []string) (output string, err error) {
	return
}

func (k *mockKubectlNoOutput) StdinRun(args []string, stdin []byte) (output string, err error) {
	return
}

type MockKubeFactory struct{}

func (m *MockKubeFactory) NewInstance(namespace string) KubeAllRunner {
	return &mockKubectl{}
}

type MockKubeFactoryNoOutput struct{}

func (m *MockKubeFactoryNoOutput) NewInstance(namespace string) KubeAllRunner {
	return &mockKubectlNoOutput{}
}
