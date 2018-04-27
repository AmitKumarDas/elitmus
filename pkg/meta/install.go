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

package meta

import (
	"fmt"
	"io/ioutil"

	"github.com/AmitKumarDas/litmus/pkg/kubectl"
	"github.com/ghodss/yaml"
)

// InstallFile type defines a yaml file path that represents an installation
// and is used for various purposes e.g. verification
type InstallFile string

// Installation represents a set of components that represent an installation
// e.g. an operator represented by its CRDs, RBACs and Deployments forms an
// installation
//
// NOTE:
//  Installation struct is accepted as a yaml file that can be used to verify.
// In addition this file allows the testing logic to take appropriate actions
// as directed in the .feature file.
type Installation struct {
	// Version of this installation, operator etc
	Version string `json:"version"`
	// Components of this installation
	Components []Component `json:"components"`
}

// Component is the information about a particular component
// e.g. a Kubernetes Deployment, or a Kubernetes Pod, etc can be
// a component in the overall installation
type Component struct {
	// Name of the component
	Name string `json:"name"`
	// Namespace of the component
	Namespace string `json:"namespace"`
	// Kind name of the component
	// e.g. pods, deployments, services, etc
	Kind string `json:"kind"`
	// APIVersion of the component
	APIVersion string `json:"apiVersion"`
	// Labels of the component that is used for filtering the components
	//
	// Following are some valid sample values for labels:
	//
	//    labels: name=app
	//    labels: name=app,env=prod
	Labels string `json:"labels"`
	// Alias provides a user understood description used for filtering the
	// components. This is a single word setting.
	//
	// NOTE:
	//  Ensure unique alias values in an installation
	//
	// DETAILS:
	//  This is the text which is typically understood by the end user. This text
	// which will be set in the installation file against a particular component.
	// Logic will filter the component based on this alias & run
	// various checks &/or actions
	Alias string `json:"alias"`
}

// unmarshal takes the raw yaml data and unmarshals it into Installation
func unmarshal(data []byte) (installation *Installation, err error) {
	installation = &Installation{}

	err = yaml.Unmarshal(data, installation)
	return
}

// load converts a verify file into an instance of *Installation
func Load(file InstallFile) (installation *Installation, err error) {
	if len(file) == 0 {
		err = fmt.Errorf("failed to load: verify file is not provided")
		return
	}

	d, err := ioutil.ReadFile(string(file))
	if err != nil {
		return
	}

	return unmarshal(d)
}

// GetMatchingPodComponent returns the pod that matches with alias
func (i *Installation) GetMatchingPodComponent(alias string) (comp Component, err error) {
	var filtered = []Component{}

	// filter the components that are pods & match with the provided alias
	for _, c := range i.Components {
		if c.Alias == alias && kubectl.IsPod(c.Kind) {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		err = fmt.Errorf("pod component not found for alias '%s'", alias)
		return
	}

	// there should be only one component that matches the alias
	if len(filtered) > 1 {
		err = fmt.Errorf("multiple components found for alias '%s': alias should be unique in an install", alias)
		return
	}

	return filtered[0], nil
}

// GetMatchingServiceComponent returns the service that matches with alias
func (i *Installation) GetMatchingServiceComponent(alias string) (comp Component, err error) {
	var filtered = []Component{}

	// filter the components that are services & match with the provided alias
	for _, c := range i.Components {
		if c.Alias == alias && kubectl.IsService(c.Kind) {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		err = fmt.Errorf("service component not found for alias '%s'", alias)
		return
	}

	// there should be only one component that matches the alias
	if len(filtered) > 1 {
		err = fmt.Errorf("multiple components found for alias '%s': alias should be unique in an install", alias)
		return
	}

	return filtered[0], nil
}

// GetJobComponent returns the job component that matches the
// provided alias
func (i *Installation) GetJobComponent(alias string) (comp Component, err error) {
	var filtered = []Component{}

	// filter the components that are jobs & match with the provided alias
	for _, c := range i.Components {
		if c.Alias == alias && kubectl.IsJob(c.Kind) {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		err = fmt.Errorf("job component not found for alias '%s'", alias)
		return
	}

	// there should be only one component that matches the alias
	if len(filtered) > 1 {
		err = fmt.Errorf("multiple components found for alias '%s': alias should be unique in an install", alias)
		return
	}

	return filtered[0], nil
}
