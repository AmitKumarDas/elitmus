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

	"github.com/AmitKumarDas/litmus/pkg/kubectl"
	"github.com/AmitKumarDas/litmus/pkg/meta"
)

// Property type defines the resource property
type Property string

const (
	// ServiceIPProperty points to service's cluster ip property
	ServiceIPProperty Property = "service-ip"
)

// Fetcher provides contract(s) i.e. method signature(s) to fetch relevant
// resource properties
type Fetcher interface {
	Fetch(alias string, property Property) (data []string, err error)
}

// KubeResourceFetch provides methods that provides methods to fetch relevant
// kubernetes resource properties
type KubeResourceFetch struct {
	// installation is the set of components that is installed on kubernetes
	installation *meta.Installation
	// kubectlFactory instance enables fetching new instance of KubeAllRunner
	kubectlFactory kubectl.KubeFactory
}

// NewKubeResourceFetch provides a new instance of KubeResourceFetch
func NewKubeResourceFetch(file meta.InstallFile) (*KubeResourceFetch, error) {
	i, err := meta.Load(file)
	if err != nil {
		return nil, err
	}

	return &KubeResourceFetch{
		installation:   i,
		kubectlFactory: kubectl.NewKubeFactory(),
	}, nil
}

// Fetch fetches a specific property of the resource identified by the alias
func (f *KubeResourceFetch) Fetch(alias string, property Property) (data []string, err error) {
	switch property {
	case ServiceIPProperty:
		return f.fetchServiceIP(alias)
	default:
		err = fmt.Errorf("property '%s' is not supported by kubernetes resource", property)
	}
	return
}

// fetchServiceIP fetches service IP of the resource identified by the alias
func (f *KubeResourceFetch) fetchServiceIP(alias string) (data []string, err error) {
	c, err := f.installation.GetMatchingServiceComponent(alias)
	if err != nil {
		return
	}

	k := f.kubectlFactory.NewInstance(c.Namespace)
	ip, err := kubectl.GetServiceIP(k, c.Name)
	if err != nil {
		return
	}

	if len(ip) == 0 {
		err = fmt.Errorf("service ip is not set for component '%#v'", c)
		return
	}

	data = append(data, ip)
	return
}
