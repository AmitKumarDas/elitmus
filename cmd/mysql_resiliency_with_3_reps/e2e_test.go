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

package main

import (
	"fmt"

	"github.com/AmitKumarDas/litmus/pkg/k8s"
	"github.com/AmitKumarDas/litmus/pkg/time"
	"github.com/AmitKumarDas/litmus/pkg/util"
	"github.com/AmitKumarDas/litmus/pkg/volume"
	"github.com/DATA-DOG/godog"
)

// volumePlacementIdentity is a type that identifies a volume
// replica based on its placement
type volumePlacementIdentity string

const (
	// FirstVPI is the identity given to the volume replica assuming
	// it to be placed on first node
	FirstVPI volumePlacementIdentity = "first"
	// SecondVPI is the identity given to the volume replica assuming
	// it to be placed on second node
	SecondVPI volumePlacementIdentity = "second"
	// ThirdVPI is the identity given to the volume replica assuming
	// it to be placed on third node
	ThirdVPI volumePlacementIdentity = "third"
)

// errorIdentity marks an error to a unique identity
type errorIdentity string

const (
	// OpenEBSConfigEI helps in finding the actual error due to openebsconfig
	// related operation
	OpenEBSConfigEI errorIdentity = "openebsconfigerr"
)

const (
	OpenebsConfigFile = "/etc/e2e/openebsconfig.yaml"
)

// volumeReplicaNode holds information about the volume replica
// along with its placement details
type volumeReplicaNode struct {
	// identifier is a unique identity for this volume replica
	// based on its placement
	identifier volumePlacementIdentity
	// name of the volume replica
	volReplicaName string
	// name of the node which schedules this volume replica
	nodeName string
}

type MySQLResiliencyWith3Reps struct {
	// appName is the name given to the mysql app
	appName string
	// volName is the name given to the volume
	volName string
	// volNodes is an list of volume replicas with replica related info
	volNodes []volumeReplicaNode
	// isLaunchOk flags if the launch of the application was successful
	isLaunchOk bool
	// kubeConnect instance helps in checking connection to kubernetes cluster
	kubeConnect k8s.KubeConnected
	// kubeRun instance enables running kubernetes specific operations
	kubeRun k8s.KubeRunner
	// operatorDeploy instance enables verification of deployment of openebs
	// operator components
	operatorDeploy volume.Deployed
	// errors hold the error(s) that occured before the suite run
	errors map[errorIdentity]error
}

func (e2e *MySQLResiliencyWith3Reps) withKubernetes() {
	// build a kubectl instance using namespace & context from environment
	// variables; will use default namespace & context if not provided
	k := k8s.NewKubectl(util.KubeNamespaceENV(), util.KubeContextENV())

	e2e.kubeConnect = k
	e2e.kubeRun = k
}

func (e2e *MySQLResiliencyWith3Reps) withOperator() {
	o, err := volume.NewOpenebsOperator(e2e.kubeRun, OpenebsConfigFile)
	if err != nil {
		e2e.errors[OpenEBSConfigEI] = err
	}
	e2e.operatorDeploy = o
}

func (e2e *MySQLResiliencyWith3Reps) iHaveAKubernetesClusterWithVolumeInstalled(volOperatorName string) (err error) {
	// checks if kubernetes cluster is available & is connected
	_, err = e2e.kubeConnect.IsConnected()
	if err != nil {
		return
	}

	if e2e.operatorDeploy == nil {
		err = fmt.Errorf("nil operator instance: possible error '%s'", e2e.errors[OpenEBSConfigEI])
		return
	}

	// checks if operator is deployed
	_, err = e2e.operatorDeploy.IsDeployed()
	if err != nil {
		return
	}

	return
}

func (e2e *MySQLResiliencyWith3Reps) iLaunchApplicationOnVolume(appName, volName string) (err error) {
	return
}

func (e2e *MySQLResiliencyWith3Reps) waitFor(duration string) (err error) {
	err = time.WaitFor(duration)
	return
}

func (e2e *MySQLResiliencyWith3Reps) verifyApplicationIsLaunchedSuccessfullyOnVolume() (err error) {
	return
}

func (e2e *MySQLResiliencyWith3Reps) applicationIsLaunchedSuccessfullyOnVolume() (err error) {
	return
}

func (e2e *MySQLResiliencyWith3Reps) verifyEachVolumeReplicaGetsAUniqueNode() (err error) {
	return
}

func (e2e *MySQLResiliencyWith3Reps) saveVolumeReplicaAndNode(identifier, replicaName, nodeName string) (err error) {
	return
}

func (e2e *MySQLResiliencyWith3Reps) iShutdownNode(identifier string) (err error) {
	return
}

func (e2e *MySQLResiliencyWith3Reps) verifyNodeIsNotAvailable(identifier string) (err error) {
	return
}

func (e2e *MySQLResiliencyWith3Reps) verifyApplicationIsRunning() (err error) {
	return
}

func (e2e *MySQLResiliencyWith3Reps) verifyVolumeReplicasAreRunning(runPercent string) (err error) {
	return
}

func (e2e *MySQLResiliencyWith3Reps) startNode(identifier string) (err error) {
	return
}

func (e2e *MySQLResiliencyWith3Reps) verifyNodeIsRunning(identifier string) (err error) {
	return
}

func FeatureContext(s *godog.Suite) {
	e2e := &MySQLResiliencyWith3Reps{
		errors: map[errorIdentity]error{},
	}

	s.BeforeSuite(e2e.withKubernetes)
	s.BeforeSuite(e2e.withOperator)

	// this associates the specs with corresponding methods of mysqlResiliencyWith3Reps
	s.Step(`^I have a kubernetes cluster with volume "([^"]*)" installed$`, e2e.iHaveAKubernetesClusterWithVolumeInstalled)
	s.Step(`^I launch application "([^"]*)" on volume "([^"]*)"$`, e2e.iLaunchApplicationOnVolume)
	s.Step(`^wait for "([^"]*)"$`, e2e.waitFor)
	s.Step(`^verify application is launched successfully on volume$`, e2e.verifyApplicationIsLaunchedSuccessfullyOnVolume)
	s.Step(`^application is launched successfully on volume$`, e2e.applicationIsLaunchedSuccessfullyOnVolume)
	s.Step(`^verify each volume replica gets a unique node$`, e2e.verifyEachVolumeReplicaGetsAUniqueNode)
	s.Step(`^save "([^"]*)" volume replica "([^"]*)" and node "([^"]*)"$`, e2e.saveVolumeReplicaAndNode)
	s.Step(`^I shutdown "([^"]*)" node$`, e2e.iShutdownNode)
	s.Step(`^verify "([^"]*)" node is not available$`, e2e.verifyNodeIsNotAvailable)
	s.Step(`^verify application is running$`, e2e.verifyApplicationIsRunning)
	s.Step(`^verify "([^"]*)" volume replicas are running$`, e2e.verifyVolumeReplicasAreRunning)
	s.Step(`^start "([^"]*)" node$`, e2e.startNode)
	s.Step(`^verify "([^"]*)" node is running$`, e2e.verifyNodeIsRunning)
}
