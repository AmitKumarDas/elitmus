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

package verify

import (
	"fmt"
	"strings"

	"github.com/AmitKumarDas/elitmus/pkg/kubectl"
	"github.com/AmitKumarDas/elitmus/pkg/meta"
	"github.com/AmitKumarDas/elitmus/pkg/util"
)

// Condition type defines a condition that can be applied against a component
// or a set of components
type Condition string

const (
	// UniqueNodeCond is a condition to check uniqueness of node
	UniqueNodeCond Condition = "is-unique-node"
	// ThreeReplicasCond is a condition to check if replica count is 3
	ThreeReplicasCond Condition = "is-three-replicas"
	// PVCBoundCond is a condition to check if PVC is bound
	PVCBoundCond Condition = "is-pvc-bound"
	// PVCUnBoundCond is a condition to check if PVC is unbound
	PVCUnBoundCond Condition = "is-pvc-unbound"
	// MultiNodeClusterCond is a condition to check if kubernetes cluster
	// has more than one node
	MultiNodeClusterCond Condition = "is-multi-node-k8s-cluster"
	// JobCompletedCond is a condition to check if job is completed
	JobCompletedCond Condition = "is-job-completed"
)

// Action type defines a action that can be applied against a component
// or a set of components
type Action string

const (
	// DeleteAnyPodAction is an action to delete any pod
	DeleteAnyPodAction Action = "delete-any-pod"
	// DeleteOldestPodAction is an action to delete the oldest pod
	DeleteOldestPodAction Action = "delete-oldest-pod"
	// CordonNodeWithOldestPodAction is an action to cordon a node that hosts
	// the oldest pod
	CordonNodeWithOldestPodAction Action = "cordon-node-with-oldest-pod"
)

// DeleteVerifier provides contract(s) i.e. method signature(s) to evaluate
// if an installation was deleted successfully
type DeleteVerifier interface {
	IsDeleted() (yes bool, err error)
}

// DeployVerifier provides contract(s) i.e. method signature(s) to evaluate
// if an installation was deployed successfully
type DeployVerifier interface {
	IsDeployed() (yes bool, err error)
}

// ConnectVerifier provides contract(s) i.e. method signature(s) to evaluate
// if a connection is possible or not
type ConnectVerifier interface {
	IsConnected() (yes bool, err error)
}

// RunVerifier provides contract(s) i.e. method signature(s) to evaluate
// if an entity is in a running state or not
type RunVerifier interface {
	IsRunning() (yes bool, err error)
}

// ConditionVerifier provides contract(s) i.e. method signature(s) to evaluate
// if specific entities passes the condition
type ConditionVerifier interface {
	IsCondition(alias string, condition Condition) (yes bool, err error)
}

// ActionVerifier provides contract(s) i.e. method signature(s) to evaluate
// if specific entities passes the action
type ActionVerifier interface {
	IsAction(alias string, action Action) (yes bool, err error)
}

// DeployRunVerifier provides contract(s) i.e. method signature(s) to
// evaluate:
//
// 1/ if an entity is deployed &,
// 2/ if the entity is running
type DeployRunVerifier interface {
	// DeployVerifier will check if the instance has been deployed or not
	DeployVerifier
	// RunVerifier will check if the instance is in a running state or not
	RunVerifier
}

// AllVerifier provides contract(s) i.e. method signature(s) to
// evaluate:
//
// - if an entity is deleted,
// - if an entity is deployed,
// - if the entity is running,
// - if the entity satisfies the provided condition &
// - if the entity satisfies the provided action
type AllVerifier interface {
	// DeleteVerifier will check if the instance has been deleted or not
	DeleteVerifier
	// DeployVerifier will check if the instance has been deployed or not
	DeployVerifier
	// RunVerifier will check if the instance is in a running state or not
	RunVerifier
	// ConditionVerifier will check if the instance satisfies the provided
	// condition
	ConditionVerifier
	// ActionVerifier will check if the instance satisfies the provided action
	ActionVerifier
}

// KubeInstallVerify provides methods that handles verification related logic of
// an installation within kubernetes e.g. application, deployment, operator, etc
type KubeInstallVerify struct {
	// installation is the set of components that determine the install
	installation *meta.Installation
}

// NewKubeInstallVerify provides a new instance of NewKubeInstallVerify based on
// the provided install file
func NewKubeInstallVerify(file meta.InstallFile) (*KubeInstallVerify, error) {
	i, err := meta.Load(file)
	if err != nil {
		return nil, err
	}

	return &KubeInstallVerify{
		installation: i,
	}, nil
}

// IsDeployed evaluates if all components of the installation are deployed
func (v *KubeInstallVerify) IsDeployed() (yes bool, err error) {
	if v.installation == nil {
		err = fmt.Errorf("failed to check IsDeployed: installation object is nil")
		return
	}

	for _, component := range v.installation.Components {
		yes, err = isComponentDeployed(component)
		if err != nil {
			break
		}
	}

	return
}

// IsDeleted evaluates if all components of the installation are deleted
func (v *KubeInstallVerify) IsDeleted() (yes bool, err error) {
	if v.installation == nil {
		err = fmt.Errorf("failed to check IsDeleted: installation object is nil")
		return
	}

	for _, component := range v.installation.Components {
		yes, err = isComponentDeleted(component)
		if err != nil {
			break
		}
	}

	return
}

// IsRunning evaluates if all components of the installation are running
func (v *KubeInstallVerify) IsRunning() (yes bool, err error) {
	if v.installation == nil {
		err = fmt.Errorf("failed to check IsRunning: installation object is nil")
		return
	}

	for _, component := range v.installation.Components {
		if component.Kind != "pod" {
			continue
		}

		yes, err = isPodComponentRunning(component)
		if err != nil {
			break
		}
	}

	return
}

// IsCondition evaluates if specific components satisfies the condition
func (v *KubeInstallVerify) IsCondition(alias string, condition Condition) (yes bool, err error) {
	switch condition {
	case UniqueNodeCond:
		return v.isEachComponentOnUniqueNode(alias)
	case ThreeReplicasCond:
		return v.hasComponentThreeReplicas(alias)
	case PVCBoundCond:
		return v.isPVCBound(alias)
	case PVCUnBoundCond:
		return v.isPVCUnBound(alias)
	case JobCompletedCond:
		return v.isJobCompleted(alias)
	default:
		err = fmt.Errorf("condition '%s' is not supported", condition)
	}
	return
}

// IsAction evaluates if specific components satisfies the action
func (v *KubeInstallVerify) IsAction(alias string, action Action) (yes bool, err error) {
	switch action {
	case DeleteAnyPodAction:
		return v.isDeleteAnyRunningPod(alias)
	case DeleteOldestPodAction:
		return v.isDeleteOldestRunningPod(alias)
	case CordonNodeWithOldestPodAction:
		return v.isCordonNodeWithOldestPod(alias)
	default:
		err = fmt.Errorf("action '%s' is not supported", action)
	}
	return
}

// isDeleteAnyPod deletes a pod based on the alias
func (v *KubeInstallVerify) isDeleteAnyRunningPod(alias string) (yes bool, err error) {
	var pods = []string{}

	c, err := v.installation.GetMatchingPodComponent(alias)
	if err != nil {
		return
	}

	if len(strings.TrimSpace(c.Labels)) == 0 {
		err = fmt.Errorf("unable to fetch component '%s' '%s': component labels are missing '%s'", c.Kind, alias)
		return
	}

	k := kubectl.New().Namespace(c.Namespace).Labels(c.Labels)
	pods, err = kubectl.GetRunningPods(k)
	if err != nil {
		return
	}

	if len(pods) == 0 {
		err = fmt.Errorf("failed to delete any running pod: pods with alias '%s' and running state are not found", alias)
		return
	}

	// delete any running pod
	k = kubectl.New().Namespace(c.Namespace)
	err = kubectl.DeletePod(k, pods[0])
	if err != nil {
		return
	}

	yes = true
	return
}

// isDeleteOldestRunningPod deletes the oldset pod based on the alias
func (v *KubeInstallVerify) isDeleteOldestRunningPod(alias string) (yes bool, err error) {
	var pod string

	c, err := v.installation.GetMatchingPodComponent(alias)
	if err != nil {
		return
	}

	// check for presence of labels
	if len(strings.TrimSpace(c.Labels)) == 0 {
		err = fmt.Errorf("failed to delete oldest running pod: component labels are missing: component '%#v': alias '%s'", c, alias)
		return
	}

	// fetch oldest running pod
	k := kubectl.New().Namespace(c.Namespace).Labels(c.Labels)
	pod, err = kubectl.GetOldestRunningPod(k)
	if err != nil {
		return
	}

	if len(pod) == 0 {
		err = fmt.Errorf("failed to delete oldest running pod: pod with running state is not found: alias '%s'", alias)
		return
	}

	// delete oldest running pod
	k = kubectl.New().Namespace(c.Namespace)
	err = kubectl.DeletePod(k, pod)
	if err != nil {
		return
	}

	yes = true
	return
}

// isCordonNodeWithOldestPod cordons the node that hosts the oldest pod. The pod
// is filtered based on the provided alias.
func (v *KubeInstallVerify) isCordonNodeWithOldestPod(alias string) (yes bool, err error) {
	var pod string

	c, err := v.installation.GetMatchingPodComponent(alias)
	if err != nil {
		return
	}

	// check for presence of labels
	if len(strings.TrimSpace(c.Labels)) == 0 {
		err = fmt.Errorf("unable to cordon node with oldest pod: component labels are missing: component '%#v': alias '%s'", c, alias)
		return
	}

	// fetch oldest running pod
	k := kubectl.New().Namespace(c.Namespace).Labels(c.Labels)
	pod, err = kubectl.GetOldestRunningPod(k)
	if err != nil {
		return
	}

	if len(pod) == 0 {
		err = fmt.Errorf("unable to cordon node with oldest pod: pod with running state is not found: alias '%s'", alias)
		return
	}

	// cordon the node that hosts this oldest pod
	k = kubectl.New().Namespace(c.Namespace)
	err = kubectl.CordonNodeWithPod(k, pod)
	if err != nil {
		return
	}

	yes = true
	return
}

// hasComponentThreeReplicas flags if a component has three replicas
func (v *KubeInstallVerify) hasComponentThreeReplicas(alias string) (yes bool, err error) {
	err = fmt.Errorf("hasComponentThreeReplicas is not implemented")
	return
}

// isJobCompleted flags if a job is completed
func (v *KubeInstallVerify) isJobCompleted(alias string) (yes bool, err error) {
	c, err := v.installation.GetMatchingPodComponent(alias)
	if err != nil {
		return
	}

	k := kubectl.New().Namespace(c.Namespace).Labels(c.Labels)
	return kubectl.AreJobPodsCompleted(k)
}

// isPVCBound flags if a PVC component is bound
func (v *KubeInstallVerify) isPVCBound(alias string) (yes bool, err error) {
	var vol string
	vol, err = v.getPVCVolume(alias)
	if err != nil {
		return
	}

	// if no vol then pvc is not bound
	if len(strings.TrimSpace(vol)) == 0 {
		err = fmt.Errorf("pvc component is not bound")
		return
	}

	yes = true
	return
}

// isPVCUnBound flags if a PVC component is unbound
func (v *KubeInstallVerify) isPVCUnBound(alias string) (yes bool, err error) {
	var vol string
	vol, err = v.getPVCVolume(alias)
	if err != nil {
		return
	}

	// if no vol then pvc is not bound
	if len(strings.TrimSpace(vol)) != 0 {
		err = fmt.Errorf("pvc component is bound")
		return
	}

	yes = true
	return
}

// isEachComponentOnUniqueNode flags if each component is placed on unique node
func (v *KubeInstallVerify) isEachComponentOnUniqueNode(alias string) (bool, error) {
	var filtered = []meta.Component{}
	var nodes = []string{}

	// filter the components based on the provided alias
	for _, c := range v.installation.Components {
		if c.Alias == alias {
			filtered = append(filtered, c)
		}
	}

	// get the node of each filtered component
	for _, f := range filtered {
		// skip for non pod components
		if !util.IsPod(f.Kind) {
			continue
		}

		// if pod then get the node on which it is running
		if len(strings.TrimSpace(f.Labels)) == 0 {
			return false, fmt.Errorf("unable to fetch component '%s' node: component labels are required", f.Kind)
		}

		k := kubectl.New().Namespace(f.Namespace).Labels(f.Labels)
		n, err := kubectl.GetPodNodes(k)
		if err != nil {
			return false, err
		}

		nodes = append(nodes, n...)
	}

	if len(nodes) == 0 {
		return false, fmt.Errorf("unable to determine component's unique node: nodes '%#v'", nodes)
	}

	// check if condition is satisfied i.e. no duplicate nodes
	exists := map[string]string{}
	for _, n := range nodes {
		if _, ok := exists[n]; ok {
			return false, nil
		}
		exists[n] = "tracked"
	}

	return true, nil
}

// getPVCVolume fetches the PVC's volume
func (v *KubeInstallVerify) getPVCVolume(alias string) (op string, err error) {
	var filtered = []meta.Component{}

	// filter the components based on the provided alias
	for _, c := range v.installation.Components {
		if c.Alias == alias {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		err = fmt.Errorf("unable to check pvc bound status: no component with alias '%s'", alias)
		return
	}

	if len(filtered) > 1 {
		err = fmt.Errorf("unable to check pvc bound status: more than one components found with alias '%s'", alias)
		return
	}

	if len(filtered[0].Name) == 0 {
		err = fmt.Errorf("unable to check pvc bound status: component name is required: '%#v'", filtered[0])
		return
	}

	if filtered[0].Kind != "pvc" {
		err = fmt.Errorf("unable to check pvc bound status: component is not a pvc resource: '%#v'", filtered[0])
		return
	}

	op, err = kubectl.New().
		Namespace(filtered[0].Namespace).
		Run([]string{"get", "pvc", filtered[0].Name, "-o", "jsonpath='{.spec.volumeName}'"})

	return
}

// isPodComponentRunning flags if a particular component is running
func isPodComponentRunning(component meta.Component) (yes bool, err error) {
	// either name or labels is required
	if len(strings.TrimSpace(component.Name)) == 0 && len(strings.TrimSpace(component.Labels)) == 0 {
		err = fmt.Errorf("unable to verify pod component running status: either name or its labels is required: component '%#v'", component)
		return
	}

	// check via name
	if len(strings.TrimSpace(component.Name)) != 0 {
		k := kubectl.New().Namespace(component.Namespace)
		return kubectl.IsPodRunning(k, component.Name)
	}

	// or check via labels
	k := kubectl.New().
		Namespace(component.Namespace).
		Labels(component.Labels)
	return kubectl.ArePodsRunning(k)
}

// isComponentDeployed flags if a particular component is deployed
func isComponentDeployed(component meta.Component) (yes bool, err error) {
	var op string

	if len(strings.TrimSpace(component.Kind)) == 0 {
		err = fmt.Errorf("unable to verify component deploy status: component kind is missing: component '%#v'", component)
		return
	}

	// either name or labels is required
	if len(strings.TrimSpace(component.Name)) == 0 && len(strings.TrimSpace(component.Labels)) == 0 {
		err = fmt.Errorf("unable to verify component deploy status: either component name or its labels is required: component '%#v'", component)
		return
	}

	// check via name
	if len(strings.TrimSpace(component.Name)) != 0 {
		op, err = kubectl.New().
			Namespace(component.Namespace).
			Run([]string{"get", component.Kind, component.Name, "-o", "jsonpath='{.metadata.name}'"})

		if err == nil && len(strings.TrimSpace(op)) != 0 {
			// yes, it is deployed
			yes = true
		}
		return
	}

	// or check via labels
	op, err = kubectl.New().
		Namespace(component.Namespace).
		Labels(component.Labels).
		Run([]string{"get", component.Kind, "-o", "jsonpath='{.items[*].metadata.name}'"})

	if err == nil && len(strings.TrimSpace(op)) != 0 {
		// yes, it is deployed
		yes = true
	}
	return
}

// isComponentDeleted flags if a particular component is deleted
func isComponentDeleted(component meta.Component) (yes bool, err error) {
	var op string

	if len(strings.TrimSpace(component.Kind)) == 0 {
		err = fmt.Errorf("unable to verify component delete status: component kind is missing: component '%#v'", component)
		return
	}

	// either name or labels is required
	if len(strings.TrimSpace(component.Name)) == 0 && len(strings.TrimSpace(component.Labels)) == 0 {
		err = fmt.Errorf("unable to verify component delete status: either component name or its labels is required: component '%#v'", component)
		return
	}

	// check via name
	if len(strings.TrimSpace(component.Name)) != 0 {
		op, err = kubectl.New().
			Namespace(component.Namespace).
			Run([]string{"get", component.Kind, component.Name})

		if err == nil {
			err = fmt.Errorf("component is not deleted: component '%#v': output '%s'", component, op)
			return
		}

		if strings.Contains(err.Error(), "(NotFound)") {
			// yes, it is deleted
			yes = true
			// We wanted to make sure that this component was deleted.
			// Hence the get operation is expected to result in NotFound error
			// from server. Now we can reset the err to nil to let the flow
			// continue
			err = nil
			return
		}

		err = fmt.Errorf("unable to verify delete status of component '%#v': output '%s'", component, op)
		return
	}

	// or check via labels
	op, err = kubectl.New().
		Namespace(component.Namespace).
		Labels(component.Labels).
		Run([]string{"get", component.Kind})

	if err != nil {
		return
	}

	if len(strings.TrimSpace(op)) == 0 || strings.Contains(op, "No resources found") {
		// yes, it is deleted
		yes = true
		return
	}

	err = fmt.Errorf("unable to verify delete status of component '%#v': output '%s'", component, op)
	return
}

// KubernetesVerify provides methods that provides methods applicable to
// kubernetes cluster
type KubernetesVerify struct{}

// NewKubeConnectionVerify provides a new instance of KubernetesVerify
func NewKubernetesVerify() *KubernetesVerify {
	return &KubernetesVerify{}
}

// IsConnected verifies if kubectl can connect to the target Kubernetes cluster
func (k *KubernetesVerify) IsConnected() (yes bool, err error) {
	_, err = kubectl.New().Run([]string{"get", "pods"})
	if err == nil {
		yes = true
	}
	return
}

// IsCondition evaluates if specific condition is satisfied or not
func (v *KubernetesVerify) IsCondition(alias string, condition Condition) (yes bool, err error) {
	switch condition {
	case MultiNodeClusterCond:
		return v.isMultiNodeCluster(alias)
	default:
		err = fmt.Errorf("condition '%s' is not supported by kubernetes verify", condition)
	}
	return
}

func (v *KubernetesVerify) isMultiNodeCluster(alias string) (yes bool, err error) {
	nodes, err := kubectl.GetAllNodeNames(kubectl.New())
	if err != nil {
		return
	}

	if len(nodes) > 1 {
		yes = true
	} else {
		err = fmt.Errorf("not a multi-node cluster: nodes '%#v'", nodes)
	}
	return
}
