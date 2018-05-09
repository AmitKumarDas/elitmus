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

package util

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
