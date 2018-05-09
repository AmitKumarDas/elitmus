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

	"github.com/AmitKumarDas/elitmus/pkg/kubectl"
	"github.com/AmitKumarDas/elitmus/pkg/meta"
	"github.com/AmitKumarDas/elitmus/pkg/time"
	"github.com/AmitKumarDas/elitmus/pkg/verify"
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
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
	OperatorMF    meta.InstallFile = "/etc/e2e/operator-verify/operator-verify.yaml"
	ApplicationMF meta.InstallFile = "/etc/e2e/application-verify/application-verify.yaml"
	VolumeMF      meta.InstallFile = "/etc/e2e/volume-verify/volume-verify.yaml"
)

const (
	ApplicationKF kubectl.KubectlFile = "/etc/e2e/application-launch/application-launch.yaml"
)

const (
	// VolumeReplicaAlias is the alias name given to a volume replica pod
	//
	// This is the text which is typically understood by the end user. This text
	// which will be set in the verify file against a particular component.
	// Verification logic will filter the component based on this alias & run
	// various checks &/or actions
	VolumeReplicaAlias string = "volume-replica"
	// VolumeDeploymentAlias is the alias name given to the volume replica deployment
	//
	// This is the text which is typically understood by the end user. This text
	// which will be set in the verify file against a particular component.
	// Verification logic will filter the component based on this alias & run
	// various checks &/or actions
	VolumeDeploymentAlias string = "volume-deployment"
)

type MySQLResiliencyWith3Reps struct {
	// appVerifier instance enables verification of application components
	appVerifier verify.DeployRunVerifier
	// volVerifier instance enables verification of volume components
	volVerifier verify.AllVerifier
	// operatorVerifier instance enables verification of operator components
	operatorVerifier verify.DeployRunVerifier
	// errors hold the previous error(s)
	errors map[errorIdentity]error
}

func (e2e *MySQLResiliencyWith3Reps) withOperatorVerifier(f *gherkin.Feature) {
	o, err := verify.NewKubeInstallVerify(OperatorMF)
	if err != nil {
		e2e.errors[OperatorVerifyFileEI] = err
		return
	}
	e2e.operatorVerifier = o
}

func (e2e *MySQLResiliencyWith3Reps) withApplicationVerifier(f *gherkin.Feature) {
	a, err := verify.NewKubeInstallVerify(ApplicationMF)
	if err != nil {
		e2e.errors[ApplicationVerifyFileEI] = err
		return
	}
	e2e.appVerifier = a
}

func (e2e *MySQLResiliencyWith3Reps) withVolumeVerifier(f *gherkin.Feature) {
	v, err := verify.NewKubeInstallVerify(VolumeMF)
	if err != nil {
		e2e.errors[VolumeVerifyFileEI] = err
		return
	}
	e2e.volVerifier = v
}

func (e2e *MySQLResiliencyWith3Reps) tearDown(f *gherkin.Feature) {
	kubectl.New().Run([]string{"delete", "-f", string(ApplicationKF)})
}

func (e2e *MySQLResiliencyWith3Reps) iHaveAKubernetesClusterWithVolumeOperatorInstalled() (err error) {
	kubeVerifier := verify.NewKubernetesVerify()
	// checks if kubernetes cluster is available & is connected
	_, err = kubeVerifier.IsConnected()
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
	_, err = kubectl.New().Run([]string{"apply", "-f", string(ApplicationKF)})
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

func (e2e *MySQLResiliencyWith3Reps) verifyThereAreThreeReplicasOfVolumeDeployment() (err error) {
	if e2e.volVerifier == nil {
		err = fmt.Errorf("nil volume verifier: possible error '%s'", e2e.errors[VolumeVerifyFileEI])
		return
	}

	// is condition satisfied
	_, err = e2e.volVerifier.IsCondition(VolumeDeploymentAlias, verify.ThreeReplicasCond)
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

	s.BeforeFeature(e2e.withOperatorVerifier)
	s.BeforeFeature(e2e.withApplicationVerifier)
	s.BeforeFeature(e2e.withVolumeVerifier)

	s.AfterFeature(e2e.tearDown)

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
	s.Step(`^verify there are three replicas of volume deployment$`, e2e.verifyThereAreThreeReplicasOfVolumeDeployment)
}
