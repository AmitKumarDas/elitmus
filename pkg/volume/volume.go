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

package volume

import (
	"fmt"
	"io/ioutil"

	"github.com/AmitKumarDas/litmus/pkg/k8s"
	"github.com/ghodss/yaml"
)

// Deployed provides the contract i.e. method signature for structure(s)
// to evaluate the structure's deployment on the target infrastructure
type Deployed interface {
	IsDeployed() (yes bool, err error)
}

// Openebs represents the components that define an openebs
// installation
type Openebs struct {
	// Version of openebs
	Version string `json:"version"`
	// Components of openebs that make openebs what it is
	Components []Component `json:"components"`
}

// Component is the information about a particular component
// e.g. an openebs component say maya api server can be described
// via this structure
type Component struct {
	// Name of the component
	Name string `json:"name"`
	// Alias name of the component
	// e.g. pods, deployments, services, etc
	Alias string `json:"alias"`
	// APIVersion of the component
	APIVersion string `json:"apiVersion"`
}

// unmarshalOpenebsConfig takes raw openebsconfig.yaml data and unmarshals it.
func unmarshalOpenebsConfig(data []byte) (openebs *Openebs, err error) {
	openebs = &Openebs{}

	err = yaml.Unmarshal(data, openebs)
	if err != nil {
		return
	}

	return
}

// loadOpenebs loads a openebsconfig.yaml file into *Openebs.
func loadOpenebs(openebsConfig string) (openebs *Openebs, err error) {
	if len(openebsConfig) == 0 {
		err = fmt.Errorf("failed to initialize openebs: openebs config is not provided")
		return
	}

	d, err := ioutil.ReadFile(openebsConfig)
	if err != nil {
		return
	}

	return unmarshalOpenebsConfig(d)
}

// OpenebsOperator provides required methods to deal with this operator
type OpenebsOperator struct {
	// openebs is the set of components that determine openebs
	openebs *Openebs
	// kubeRunner enables execution of kubernetes operations
	kubeRunner k8s.KubeRunner
}

// NewOpenebsOperator provides a new instance of OpenebsOperator
// based on the provided kubernetes runner
func NewOpenebsOperator(runner k8s.KubeRunner, openebsConfig string) (oo *OpenebsOperator, err error) {
	o, err := loadOpenebs(openebsConfig)
	if err != nil {
		return
	}

	oo = &OpenebsOperator{
		kubeRunner: runner,
		openebs:    o,
	}

	return
}

// IsDeployed evaluates if all components of the operator are deployed
func (o *OpenebsOperator) IsDeployed() (yes bool, err error) {
	if o.openebs == nil {
		err = fmt.Errorf("failed to check IsDeployed: openebs config is not initialized")
		return
	}

	for _, component := range o.openebs.Components {
		yes, err = o.IsComponentDeployed(component)
		if err != nil {
			break
		}
	}

	return
}

// IsComponentDeployed flags if a particular component is deployed
func (o *OpenebsOperator) IsComponentDeployed(component Component) (yes bool, err error) {
	_, err = o.kubeRunner.Run([]string{"get", component.Alias, component.Name}, "")
	if err == nil {
		yes = true
	}
	return
}
