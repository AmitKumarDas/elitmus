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
	"fmt"
	"strings"

	"github.com/AmitKumarDas/elitmus/pkg/exec"
	"github.com/AmitKumarDas/elitmus/pkg/util"
)

// KubectlFile is the type to hold various yaml file paths
// that can be applied by kubectl
type KubectlFile string

// TemplatedKubectlFile is the type to hold various yaml file paths that are
// run through go template before be applied by kubectl
type TemplatedKubectlFile string

const (
	// KubectlPath is the expected location where kubectl executable may be found
	KubectlPath = "/usr/local/bin/kubectl"
)

const (
	// DefaultLitmusNamespace is the default namespace where kubectl operations
	// will be executed
	DefaultLitmusNamespace = "litmus"
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

// KubeRunner interface provides the contract i.e. method signature to
// invoke commands at kubernetes cluster
type KubeRunner interface {
	// Run executes the kubectl command
	Run(args []string) (output string, err error)
}

// KubeStdinRunner interface provides the contract i.e. method signature to
// invoke commands at kubernetes cluster
type KubeStdinRunner interface {
	// StdinRun executes the kubectl command
	StdinRun(args []string, stdin []byte) (output string, err error)
}

// KubeAllRunner interface composes various kube runner interfaces
type KubeAllRunner interface {
	// KubeRunner interface that provides contracts to invoke commands
	// against the kubernetes cluster
	KubeRunner
	// KubeStdinRunner interface that provides contracts to invoke
	// commands against the kubernetes cluster
	KubeStdinRunner
}

// KubeFactory provides contract(s) to get a new instance of KubeAllRunner
type KubeFactory interface {
	NewInstance(namespace string) KubeAllRunner
}

// kubeFactory implements KubeFactory interface
type kubeFactory struct{}

// NewKubeFactory returns a new instance of kubeFactory
func NewKubeFactory() KubeFactory {
	return &kubeFactory{}
}

// NewInstance returns a new instance of KubeAllRunner
func (k *kubeFactory) NewInstance(namespace string) KubeAllRunner {
	return New().Namespace(namespace)
}

// Kubectl holds the properties required to execute any kubectl command.
// Kubectl is an implementation of following interfaces:
// 1. KubeRunner
type Kubectl struct {
	// namespace where this kubectl command will be run
	namespace string
	// labels to be used during kubectl execution
	labels string
	// context where this kubectl command will be run
	context string
	// args are provided to kubectl command during its run
	args []string
	// executor does actual kubectl execution
	executor exec.AllExecutor
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

// New returns a new instance of kubectl based on defaults
func New() *Kubectl {
	return &Kubectl{
		namespace: DefaultLitmusNamespace,
		executor:  exec.NewShellExec(GetKubectlPath()),
	}
}

// Namespace sets the namespace to be used during kubectl run
func (k *Kubectl) Namespace(namespace string) *Kubectl {
	if len(namespace) == 0 {
		return k
	}
	k.namespace = namespace
	return k
}

// Labels sets the labels to be used during kubectl run
func (k *Kubectl) Labels(labels string) *Kubectl {
	k.labels = labels
	return k
}

// Context sets the context to be used during kubectl run
func (k *Kubectl) Context(context string) *Kubectl {
	k.context = context
	return k
}

// Args sets the args to be used during kubectl run
func (k *Kubectl) Args(args []string) *Kubectl {
	k.args = args
	return k
}

// Run will execute the kubectl command & provide output or error
func (k *Kubectl) Run(args []string) (output string, err error) {
	k.args = kubectlArgs(args, k.namespace, k.context, k.labels)

	output, err = k.executor.Execute(k.args)
	return
}

// StdinRun will execute the kubectl command & provide output or error
func (k *Kubectl) StdinRun(args []string, stdin []byte) (output string, err error) {
	k.args = kubectlArgs(args, k.namespace, k.context, k.labels)

	output, err = k.executor.StdinExecute(k.args, stdin)
	return
}

// IsPod flags if the provided kind is a kubernetes pod or is related
// to a pod
func IsPod(kind string) (yes bool) {
	switch kind {
	case "po", "pod", "pods", "deploy", "deployment", "deployments", "job", "jobs", "sts", "statefulset", "statefulsets", "ds", "daemonset", "daemonsets":
		yes = true
	default:
		yes = false
	}

	return
}

// IsJob flags if the provided kind is a kubernetes job
func IsJob(kind string) (yes bool) {
	switch kind {
	case "job", "jobs":
		yes = true
	default:
		yes = false
	}

	return
}

// IsService flags if the provided kind is a kubernetes service
func IsService(kind string) (yes bool) {
	switch kind {
	case "svc", "service", "services":
		yes = true
	default:
		yes = false
	}

	return
}

// ArePodsRunning returns true if all the pod(s) are running, false otherwise
//
// An example of kubectl get pods & its state:
//  This makes use of status.containerStatuses[*].state.*.reason
//
// $ kubectl get po -n kube-system --selector=k8s-app=kube-dns -o jsonpath='{.items[*].status.containerStatuses[*].state.*.reason}'
// CrashLoopBackOff CrashLoopBackOff
//
// Another example of kubectl get pods & its state:
//  This makes use of status.containerStatuses[*].ready
//
// $ kubectl get pods -n kube-system --selector=k8s-app=kube-dns -o jsonpath='{.items[*].status.containerStatuses[*].ready}'
// true false true
func ArePodsRunning(k KubeRunner) (yes bool, err error) {
	isReady, err := k.Run([]string{"get", "pods", "-o", "jsonpath='{.items[*].status.containerStatuses[*].ready}'"})
	if err != nil {
		return
	}

	// split the output by space
	isReadyArr := strings.Split(strings.TrimSpace(isReady), " ")

	if contains(isReadyArr, "false") {
		err = fmt.Errorf("pod(s) are not running: '%#v' '%#v'", k, isReadyArr)
		return
	}

	// double check
	if contains(isReadyArr, "true") {
		yes = true
	} else {
		err = fmt.Errorf("status of pod(s) could not be determined: '%#v' '%#v'", k, isReadyArr)
	}

	return
}

// IsPodRunning returns true if the specified pod is running, false otherwise
//
// An example of kubectl get pods & its state:
//  This makes use of status.containerStatuses[*].state.*.reason
//
// $ kubectl get po -n kube-system my-pod -o jsonpath='{.status.containerStatuses[*].state.*.reason}'
// CrashLoopBackOff CrashLoopBackOff
//
// Another example of kubectl get pods & its state:
//  This makes use of status.containerStatuses[*].ready
//
// $ kubectl get pods -n kube-system my-pod -o jsonpath='{.status.containerStatuses[*].ready}'
// true false true
func IsPodRunning(k KubeRunner, name string) (yes bool, err error) {
	if len(name) == 0 {
		err = fmt.Errorf("unable to determine pod running status: pod name is missing")
		return
	}

	isReady, err := k.Run([]string{"get", "pods", name, "-o", "jsonpath='{.status.containerStatuses[*].ready}'"})
	if err != nil {
		return
	}

	// split the output by space
	isReadyArr := strings.Split(strings.TrimSpace(isReady), " ")

	if contains(isReadyArr, "false") {
		err = fmt.Errorf("pod '%s' is not running: '%#v'", name, isReadyArr)
		return
	}

	// double check
	if contains(isReadyArr, "true") {
		yes = true
	} else {
		err = fmt.Errorf("status of pod '%s' could not be determined: received output '%#v'", name, isReadyArr)
	}

	return
}

// GetPodNodes fetches the nodes that hosts the pods. Pods are referred to
// via the provided labels
func GetPodNodes(k KubeRunner) (nodes []string, err error) {
	n, err := k.Run([]string{"get", "pods", "-o", "jsonpath='{.items[*].spec.nodeName}'"})
	if err != nil {
		return
	}

	// split the output by space
	nodes = strings.Split(strings.TrimSpace(n), " ")
	return
}

// GetAllPodNames fetches the names of all the pods based on the labels & namespace set
// against the KubeRunner
func GetAllPodNames(k KubeRunner) (pods []string, err error) {
	p, err := k.Run([]string{"get", "pods", "-o", "jsonpath='{.items[*].metadata.name}'"})
	if err != nil {
		return
	}

	// split the output by space
	pods = strings.Split(strings.TrimSpace(p), " ")
	return
}

// GetAllNodeNames fetches the names of all the nodes registered to the cluster
func GetAllNodeNames(k KubeRunner) (nodes []string, err error) {
	op, err := k.Run([]string{"get", "nodes", "-o", "jsonpath='{.items[*].metadata.name}'"})
	if err != nil {
		return
	}

	// split the output by space
	nodes = strings.Split(strings.TrimSpace(op), " ")
	return
}

// GetRunningPods fetches the pods which are running based on the provided labels
//
// Sample code to do this:
//
// $ JSONPATH='{range .items[*]}{@.metadata.name}::{@.status.containerStatuses[*].ready}::::{end}'  && kubectl get po -n kube-system -o jsonpath="$JSONPATH"
// kube-addon-manager-amit-thinkpad-l470::true::::kube-dns-54cccfbdf8-q7v2c::false false true::::kubernetes-dashboard-77d8b98585-cwbjq::false::::storage-provisioner::true::::tiller-deploy-5b48764ff7-g9qz7::true::::
func GetRunningPods(k KubeRunner) (pods []string, err error) {
	// fetch pods
	o, err := k.Run([]string{"get", "pods", "-o", "jsonpath='{range .items[*]}{@.metadata.name}::{@.status.containerStatuses[*].ready}::::{end}'"})
	if err != nil {
		return
	}

	// split the output by the splitter used in above command
	firstSplit := strings.Split(strings.TrimSpace(o), "::::")
	for _, fs := range firstSplit {
		if len(fs) == 0 {
			continue
		}

		secondSplit := strings.Split(strings.TrimSpace(fs), "::")
		// ignore if pod is not running
		if strings.Contains(secondSplit[1], "false") {
			continue
		}

		// add the running pod to the list
		if strings.Contains(secondSplit[1], "true") {
			pods = append(pods, secondSplit[0])
		}
	}

	return
}

// GetOldestRunningPod fetches the oldest running pod based on the provided labels
// and sorted based on their age
//
// Sample code to do this:
//
// $ JSONPATH='{range .items[*]}{@.metadata.name}::{@.status.containerStatuses[*].ready}::::{end}'  && kubectl get po -n kube-system --sort-by=.metadata.creationTimestamp -o jsonpath="$JSONPATH"
// kube-addon-manager-amit-thinkpad-l470::true::::kube-dns-54cccfbdf8-q7v2c::false false true::::kubernetes-dashboard-77d8b98585-cwbjq::false::::storage-provisioner::true::::tiller-deploy-5b48764ff7-g9qz7::true::::
func GetOldestRunningPod(k KubeRunner) (pod string, err error) {
	// fetch pods sorted by creation timestamp
	o, err := k.Run([]string{"get", "pods", "--sort-by=.metadata.creationTimestamp", "-o", "jsonpath='{range .items[*]}{@.metadata.name}::{@.status.containerStatuses[*].ready}::::{end}'"})
	if err != nil {
		return
	}

	// split the output by the splitter used in above command
	firstSplit := strings.Split(strings.TrimSpace(o), "::::")
	for _, fs := range firstSplit {
		if len(fs) == 0 {
			continue
		}

		secondSplit := strings.Split(strings.TrimSpace(fs), "::")
		// ignore if pod is not running
		if strings.Contains(secondSplit[1], "false") {
			continue
		}

		// return the first running pod
		if strings.Contains(secondSplit[1], "true") {
			pod = secondSplit[0]
			return
		}
	}

	return
}

// DeletePod deletes the specified pod
func DeletePod(k KubeRunner, name string) (err error) {
	_, err = k.Run([]string{"delete", "pods", name})
	return
}

// CordonNodeWithPod cordons the node that host the specified pod
func CordonNodeWithPod(k KubeRunner, pod string) (err error) {
	op, err := k.Run([]string{"get", "pods", pod, "-o", "jsonpath='{.spec.nodeName}'"})
	if err != nil {
		return
	}

	node := strings.TrimSpace(op)
	if len(node) == 0 {
		err = fmt.Errorf("unable to cordon node: node not found for pod '%s'", pod)
		return
	}

	_, err = k.Run([]string{"cordon", node})
	return
}

// GetServiceIP gets the cluster IP address of the service
func GetServiceIP(k KubeRunner, service string) (ip string, err error) {
	return k.Run([]string{"get", "services", service, "-o", "jsonpath='{.spec.clusterIP}'"})
}

// IsJobCompleted flags if the job is completed
func IsJobCompleted(k KubeRunner, name string) (yes bool, err error) {
	op, _ := k.Run([]string{"get", "jobs", name, "-o", "jsonpath='{.status.succeeded}'"})
	if op == "1" {
		yes = true
	} else {
		err = fmt.Errorf("job is not completed: command '%#v'", k)
	}

	return
}

// AreJobPodsCompleted flags if the job's pods are completed
//
// NOTE: This expects the KubeRunner with appropriate label selector
func AreJobPodsCompleted(k KubeRunner) (yes bool, err error) {
	op, err := k.Run([]string{"get", "pods", "-o", "jsonpath='{.items[*].status.phase}'"})
	if err != nil {
		return
	}

	// split the output by space
	states := strings.Split(strings.ToLower(op), " ")
	for _, state := range states {
		if state != "succeeded" {
			err = fmt.Errorf("job pod is not completed '%#v': output '%#v'", k, states)
			return
		}
	}

	yes = true
	return
}

// UnCordonAllNodes uncordons all the nodes in the cluster
func UnCordonAllNodes(ignoreError bool) (err error) {
	nodes, err := GetAllNodeNames(New())
	if err != nil {
		return
	}

	for _, n := range nodes {
		_, err = New().Run([]string{"uncordon", n})
		if !ignoreError && err != nil {
			return
		}
	}

	return
}

// ApplyStdIn does a kubectl apply from stdin
func ApplyStdIn(stdin []byte) (err error) {
	_, err = New().StdinRun([]string{"apply", "-f", "-"}, stdin)
	return
}

// contains verifies if a specific element is present in the provided array
func contains(s []string, e string) bool {
	for _, a := range s {
		a = strings.TrimSpace(a)
		if a == e {
			return true
		}
	}
	return false
}
