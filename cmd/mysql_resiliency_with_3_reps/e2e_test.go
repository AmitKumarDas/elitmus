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

	"github.com/AmitKumarDas/litmus/pkg/kubectl"
	"github.com/AmitKumarDas/litmus/pkg/time"
	"github.com/AmitKumarDas/litmus/pkg/util"
	"github.com/AmitKumarDas/litmus/pkg/verify"
	"github.com/DATA-DOG/godog"
)

// errorIdentity marks an error to a unique identity
type errorIdentity string

const (
	// OperatorVerifyFileEI helps in finding the actual error due to operator
	// verify file related operation
	OperatorVerifyFileEI errorIdentity = "operator-verify-file-err"
	// ApplicationVerifyFileEI helps in finding the actual error due to
	// application verify file related operation
	ApplicationVerifyFileEI errorIdentity = "application-verify-file-err"
	// VolumeVerifyFileEI helps in finding the actual error due to
	// volume verify file related operation
	VolumeVerifyFileEI errorIdentity = "volume-verify-file-err"
)

const (
	OperatorMF    verify.VerifyFile = "/etc/e2e/operator-verify/operator-verify.yaml"
	ApplicationMF verify.VerifyFile = "/etc/e2e/application-verify/application-verify.yaml"
	VolumeMF      verify.VerifyFile = "/etc/e2e/volume-verify/volume-verify.yaml"
)

const (
	ApplicationKF kubectl.KubectlFile = "/etc/e2e/application-launch/application-launch.yaml"
)

const (
	// VolumeReplicaAlias is the alias name given to a volume replica
	VolumeReplicaAlias string = "volume-replica"
)

type MySQLResiliencyWith3Reps struct {
	// kubectl instance enables running kubectl operations
	kubectl kubectl.KubeRunner
	// kubeConnectionVerifier instance helps in verifying connection to kubernetes
	// cluster
	kubeConnectionVerifier verify.ConnectVerifier
	// appVerifier instance enables verification of application components
	appVerifier verify.DeployRunVerifier
	// volVerifier instance enables verification of volume components
	volVerifier verify.AllVerifier
	// operatorVerifier instance enables verification of operator components
	operatorVerifier verify.DeployRunVerifier
	// lastError holds the previos error
	lastError error
	// errors hold the previous error(s)
	errors map[errorIdentity]error
}

func (e2e *MySQLResiliencyWith3Reps) withKubernetes() {
	// build a kubectl instance using namespace & context from environment
	// variables; will use default namespace & context if not provided
	k := kubectl.NewKubectl(util.KubeNamespaceENV(), util.KubeContextENV())
	e2e.kubectl = k
}

func (e2e *MySQLResiliencyWith3Reps) withKubeConnectionVerifier() {
	k := verify.NewKubeConnectionVerify(e2e.kubectl)
	e2e.kubeConnectionVerifier = k
}

func (e2e *MySQLResiliencyWith3Reps) withOperatorVerifier() {
	o, err := verify.NewKubeInstallVerify(e2e.kubectl, OperatorMF)
	if err != nil {
		e2e.lastError = err
		e2e.errors[OperatorVerifyFileEI] = err
		return
	}
	e2e.operatorVerifier = o
}

func (e2e *MySQLResiliencyWith3Reps) withApplicationVerifier() {
	a, err := verify.NewKubeInstallVerify(e2e.kubectl, ApplicationMF)
	if err != nil {
		e2e.lastError = err
		e2e.errors[ApplicationVerifyFileEI] = err
		return
	}
	e2e.appVerifier = a
}

func (e2e *MySQLResiliencyWith3Reps) withVolumeVerifier() {
	v, err := verify.NewKubeInstallVerify(e2e.kubectl, VolumeMF)
	if err != nil {
		e2e.lastError = err
		e2e.errors[VolumeVerifyFileEI] = err
		return
	}
	e2e.volVerifier = v
}

func (e2e *MySQLResiliencyWith3Reps) iHaveAKubernetesClusterWithVolumeOperatorInstalled() (err error) {
	if e2e.kubeConnectionVerifier == nil {
		err = fmt.Errorf("nil kubernetes connection verifier")
		return
	}

	// checks if kubernetes cluster is available & is connected
	_, err = e2e.kubeConnectionVerifier.IsConnected()
	if err != nil {
		return
	}

	if e2e.operatorVerifier == nil {
		err = fmt.Errorf("nil operator verifier: possible error '%s'", e2e.errors[OperatorVerifyFileEI])
		return
	}

	// checks if operator is deployed
	_, err = e2e.operatorVerifier.IsDeployed()
	if err != nil {
		return
	}

	// checks if operator is running
	_, err = e2e.operatorVerifier.IsRunning()

	return
}

func (e2e *MySQLResiliencyWith3Reps) iLaunchMysqlApplicationOnVolume() (err error) {
	// do a kubectl apply of application yaml
	_, err = e2e.kubectl.Run([]string{"apply", "-f", string(ApplicationKF)}, "", "")
	return
}

func (e2e *MySQLResiliencyWith3Reps) waitFor(duration string) (err error) {
	err = time.WaitFor(duration)
	return
}

func (e2e *MySQLResiliencyWith3Reps) verifyMysqlApplicationIsRunning() (err error) {
	if e2e.appVerifier == nil {
		err = fmt.Errorf("nil application verifier: possible error '%s'", e2e.errors[ApplicationVerifyFileEI])
		return
	}

	// is application deployed
	_, err = e2e.appVerifier.IsDeployed()
	if err != nil {
		return
	}

	// is application running
	_, err = e2e.appVerifier.IsRunning()
	return
}

func (e2e *MySQLResiliencyWith3Reps) verifyAllVolumeReplicasAreRunning() (err error) {
	if e2e.volVerifier == nil {
		err = fmt.Errorf("nil volume verifier: possible error '%s'", e2e.errors[VolumeVerifyFileEI])
		return
	}

	// is volume deployed
	_, err = e2e.volVerifier.IsDeployed()
	if err != nil {
		return
	}

	// is volume running
	_, err = e2e.volVerifier.IsRunning()
	return
}

func (e2e *MySQLResiliencyWith3Reps) verifyMysqlApplicationIsLaunchedSuccessfullyOnVolume() (err error) {
	// check if application is running
	err = e2e.verifyMysqlApplicationIsRunning()
	if err != nil {
		return
	}

	// check if volume is running
	return e2e.verifyAllVolumeReplicasAreRunning()
}

func (e2e *MySQLResiliencyWith3Reps) mysqlApplicationIsLaunchedSuccessfullyOnVolume() (err error) {
	return e2e.verifyMysqlApplicationIsLaunchedSuccessfullyOnVolume()
}

func (e2e *MySQLResiliencyWith3Reps) verifyEachVolumeReplicaGetsAUniqueNode() (err error) {
	if e2e.volVerifier == nil {
		err = fmt.Errorf("nil volume verifier: possible error '%s'", e2e.errors[VolumeVerifyFileEI])
		return
	}

	// is condition satisfied
	_, err = e2e.volVerifier.IsCondition(VolumeReplicaAlias, verify.UniqueNodeCond)
	return
}

func (e2e *MySQLResiliencyWith3Reps) iDeleteAVolumeReplica() (err error) {
	if e2e.volVerifier == nil {
		err = fmt.Errorf("nil volume verifier: possible error '%s'", e2e.errors[VolumeVerifyFileEI])
		return
	}

	// is action satisfied
	_, err = e2e.volVerifier.IsAction(VolumeReplicaAlias, verify.DeleteAnyPodAction)
	return
}

func (e2e *MySQLResiliencyWith3Reps) iDeleteAnotherVolumeReplica() (err error) {
	if e2e.volVerifier == nil {
		err = fmt.Errorf("nil volume verifier: possible error '%s'", e2e.errors[VolumeVerifyFileEI])
		return
	}

	// is action satisfied
	_, err = e2e.volVerifier.IsAction(VolumeReplicaAlias, verify.DeleteOldestPodAction)
	return
}

func FeatureContext(s *godog.Suite) {
	e2e := &MySQLResiliencyWith3Reps{
		errors: map[errorIdentity]error{},
	}

	s.BeforeSuite(e2e.withKubernetes)
	s.BeforeSuite(e2e.withKubeConnectionVerifier)
	s.BeforeSuite(e2e.withOperatorVerifier)
	s.BeforeSuite(e2e.withApplicationVerifier)
	s.BeforeSuite(e2e.withVolumeVerifier)

	// this associates the specs with corresponding methods of mysqlResiliencyWith3Reps
	s.Step(`^I have a kubernetes cluster with volume operator installed$`, e2e.iHaveAKubernetesClusterWithVolumeOperatorInstalled)
	s.Step(`^I launch mysql application on volume$`, e2e.iLaunchMysqlApplicationOnVolume)
	s.Step(`^wait for "([^"]*)"$`, e2e.waitFor)
	s.Step(`^verify mysql application is launched successfully on volume$`, e2e.verifyMysqlApplicationIsLaunchedSuccessfullyOnVolume)
	s.Step(`^mysql application is launched successfully on volume$`, e2e.mysqlApplicationIsLaunchedSuccessfullyOnVolume)
	s.Step(`^verify each volume replica gets a unique node$`, e2e.verifyEachVolumeReplicaGetsAUniqueNode)
	s.Step(`^verify mysql application is running$`, e2e.verifyMysqlApplicationIsRunning)
	s.Step(`^verify all volume replicas are running$`, e2e.verifyAllVolumeReplicasAreRunning)
	s.Step(`^I delete a volume replica$`, e2e.iDeleteAVolumeReplica)
	s.Step(`^I delete another volume replica$`, e2e.iDeleteAnotherVolumeReplica)
}
