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

package k8s

import (
	"fmt"
	"strings"

	"github.com/AmitKumarDas/litmus/pkg/exec"
	"github.com/AmitKumarDas/litmus/pkg/util"
)

const (
	// KubectlPath is the expected location where kubectl executable may be found
	KubectlPath = "/usr/local/bin/kubectl"
)

// kubectlArgs builds the arguments required to execute any kubectl command
//
// This has been borrowed from https://github.com/CanopyTax/ckube
func kubectlArgs(args []string, namespace string, context string, labels string) []string {
	if len(namespace) != 0 {
		args = append(args, fmt.Sprintf("--namespace=%v", strings.TrimSpace(namespace)))
	}
	if len(context) != 0 {
		args = append(args, fmt.Sprintf("--context=%v", strings.TrimSpace(context)))
	}
	if len(labels) != 0 {
		args = append(args, fmt.Sprintf("--selector=%v", strings.TrimSpace(labels)))
	}
	return args
}

// KubeConnected interface provides the contract i.e. method signature to
// check connection to kubernetes cluster
type KubeConnected interface {
	IsConnected() (yes bool, err error)
}

// KubeRunner interface provides the contract i.e. method signature to
// invoke commands at kubernetes cluster
type KubeRunner interface {
	Run(args []string, labels string) (output string, err error)
}

// Kubectl holds the properties required to execute any kubectl command.
// Kubectl is an implementation of following interfaces:
// 1. KubeRunner
// 2. KubeConnected
type Kubectl struct {
	// namespace where this kubectl command will be run
	namespace string
	// context where this kubectl command will be run
	context string
	// executor does actual kubectl execution
	executor exec.Executor
}

// GetKubectlPath gets the location where kubectl executable is
// expected to be present
func GetKubectlPath() string {
	// get from environment variable
	kpath := util.KubectlPathENV()
	if len(kpath) == 0 {
		// else use the constant
		kpath = KubectlPath
	}

	return kpath
}

// NewKubectl will return a new instance of kubectl based on the provided
// information i.e. namespace, context.
func NewKubectl(namespace string, context string) *Kubectl {
	return &Kubectl{
		namespace: namespace,
		context:   context,
		executor:  exec.NewShellExec(GetKubectlPath()),
	}
}

// Run will execute the kubectl command & provide output or error
func (k *Kubectl) Run(args []string, labels string) (output string, err error) {
	kargs := kubectlArgs(args, k.namespace, k.context, labels)

	output, err = k.executor.Output(kargs)
	return
}

// IsConnected verifies if kubectl can connect to the target Kubernetes cluster
func (k *Kubectl) IsConnected() (yes bool, err error) {
	kargs := kubectlArgs([]string{"get", "pods"}, k.namespace, k.context, "")

	_, err = k.executor.Output(kargs)
	if err == nil {
		yes = true
	}

	return
}
