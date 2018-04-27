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
	"bytes"
	"fmt"
	"text/template"

	"github.com/AmitKumarDas/litmus/pkg/exec"
	"github.com/AmitKumarDas/litmus/pkg/fetch"
	"github.com/AmitKumarDas/litmus/pkg/kubectl"
	"github.com/AmitKumarDas/litmus/pkg/meta"
	"github.com/AmitKumarDas/litmus/pkg/time"
	"github.com/AmitKumarDas/litmus/pkg/verify"
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
)

// errorIdentity is a type to set error identities
type errorIdentity string

const (
	// OperatorVerifyFileEI stores the actual error during load of volume
	// operator verify file
	OperatorVerifyFileEI errorIdentity = "operator-verify-file-err"
	// ApplicationVerifyFileEI stores the actual error during load of application
	// verify file
	ApplicationVerifyFileEI errorIdentity = "application-verify-file-err"
	// VolumeVerifyFileEI stores the actual error during load of volume verify
	// file
	VolumeVerifyFileEI errorIdentity = "volume-verify-file-err"
	// AppClientConfigVerifyFileEI stores the actual error during load of application
	// client config verify file
	AppClientConfigVerifyFileEI errorIdentity = "application-client-config-verify-file-err"
)

const (
	// OperatorIF enables litmus to run checks & actions based on this volume
	// operator verify file
	OperatorIF meta.InstallFile = "/etc/e2e/operator-verify/operator-verify.yaml"
	// ApplicationIF enables litmus to run checks & actions based on this application
	// verify file
	ApplicationIF meta.InstallFile = "/etc/e2e/app-verify/app-verify.yaml"
	// AppClientConfigIF enables litmus to run checks & actions based on this
	// application client config verify file
	AppClientConfigIF meta.InstallFile = "/etc/e2e/app-client-config-verify/app-client-config-verify.yaml"
	// AppClientJobIF enbles litmus to run checks & actions based on this
	// application client job verify file
	AppClientJobIF meta.InstallFile = "/etc/e2e/app-client-job-verify/app-client-job-verify.yaml"
	// VolumeIF enables litmus to run checks & actions based on this volume verify
	// file
	VolumeIF meta.InstallFile = "/etc/e2e/volume-verify/volume-verify.yaml"
)

const (
	// ApplicationKF is the file to launch the application. This file is applied
	// via kubectl.
	ApplicationKF kubectl.KubectlFile = "/etc/e2e/app-launch/app-launch.yaml"
	// AppClientPutKF is the file to launch a job responsible to put data into app
	// This file is applied via kubectl.
	AppClientPutKF kubectl.KubectlFile = "/etc/e2e/app-client-put/app-client-put-job.yaml"
	// AppClientGetKF is the file to launch a job responsible to get data into app
	// This file is applied via kubectl.
	AppClientGetKF kubectl.KubectlFile = "/etc/e2e/app-client-get/app-client-get-job.yaml"
)

const (
	// AppClientConfigTKF is the file to be deployed before launching the application.
	// This file is templated & needs to be executed via go template & then
	// applied via kubectl.
	AppClientConfigTKF kubectl.TemplatedKubectlFile = "/etc/e2e/app-client-configs/app-client-configs.yaml"
)

const (
	// PVCAlias is the alias name given to the application's pvc
	//
	// This is the text which is typically understood by the end user. This text
	// which will be set in the verify file against a particular component.
	// Verification logic will filter the component based on this alias & run
	// various checks &/or actions
	PVCAlias string = "pvc"
	// AppPodAlias is the alias name given to the application's pod
	//
	// This is the text which is typically understood by the end user. This text
	// which will be set in the verify file against a particular component.
	// Verification logic will filter the component based on this alias & run
	// various checks &/or actions
	AppPodAlias string = "app-pod"
	// AppServiceAlias is the alias name given to the application's service
	//
	// This is the text which is typically understood by the end user. This text
	// which will be set in the verify file against a particular component.
	// Verification logic will filter the component based on this alias & run
	// various checks &/or actions
	AppServiceAlias string = "app-service"
	// GetJobAlias is the alias name given to the application client's get job
	//
	// This is the text which is typically understood by the end user. This text
	// which will be set in the verify file against a particular component.
	// Verification logic will filter the component based on this alias & run
	// various checks &/or actions
	GetJobAlias string = "get-job"
	// PutJobAlias is the alias name given to the application client's put job
	//
	// This is the text which is typically understood by the end user. This text
	// which will be set in the verify file against a particular component.
	// Verification logic will filter the component based on this alias & run
	// various checks &/or actions
	PutJobAlias string = "put-job"
	// NAAlias is the alias to be used when providing alias is not required. This
	// is only required to satisfy method signature.
	NAAlias string = "not-applicable"
)

type HAOnMinio struct {
	// appVerifier instance enables verification of application components
	appVerifier verify.AllVerifier
	// volVerifier instance enables verification of persistent volume components
	volVerifier verify.AllVerifier
	// appcJobVerifier instance enables verification of application client's job
	// completions
	appcJobVerifier verify.AllVerifier
	// operatorVerifier instance enables verification of volume operator components
	operatorVerifier verify.DeployRunVerifier
	// errors hold the previous error(s)
	errors map[errorIdentity]error
}

func (e2e *HAOnMinio) withOperatorVerifier(f *gherkin.Feature) {
	o, err := verify.NewKubeInstallVerify(OperatorIF)
	if err != nil {
		e2e.errors[OperatorVerifyFileEI] = err
		return
	}
	e2e.operatorVerifier = o
}

func (e2e *HAOnMinio) withApplicationVerifier(f *gherkin.Feature) {
	a, err := verify.NewKubeInstallVerify(ApplicationIF)
	if err != nil {
		e2e.errors[ApplicationVerifyFileEI] = err
		return
	}
	e2e.appVerifier = a
}

func (e2e *HAOnMinio) withVolumeVerifier(f *gherkin.Feature) {
	v, err := verify.NewKubeInstallVerify(VolumeIF)
	if err != nil {
		e2e.errors[VolumeVerifyFileEI] = err
		return
	}
	e2e.volVerifier = v
}

// tearDown will delete the resources that were applied during the course of
// test run
func (e2e *HAOnMinio) tearDown(f *gherkin.Feature) {
	kubectl.UnCordonAllNodes(true)
	kubectl.New().Run([]string{"delete", "-f", string(ApplicationKF)})
	kubectl.New().Run([]string{"delete", "-f", string(AppClientGetKF)})
	kubectl.New().Run([]string{"delete", "-f", string(AppClientPutKF)})
	kubectl.New().Run([]string{"delete", "-f", string(AppClientConfigTKF)})
}

func (e2e *HAOnMinio) iHaveAKubernetesMultiNodeCluster() (err error) {
	kubeVerifier := verify.NewKubernetesVerify()
	// checks if kubernetes cluster is available & is connected
	_, err = kubeVerifier.IsConnected()
	if err != nil {
		return
	}

	// is this multi node cluster?
	_, err = kubeVerifier.IsCondition(NAAlias, verify.MultiNodeClusterCond)
	return
}

func (e2e *HAOnMinio) thisClusterHasVolumeOperatorInstalled() (err error) {
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

func (e2e *HAOnMinio) waitFor(duration string) (err error) {
	err = time.WaitFor(duration)
	return
}

func (e2e *HAOnMinio) iLaunchMinioApplicationOnVolume() (err error) {
	// do a kubectl apply of application yaml
	_, err = kubectl.New().Run([]string{"apply", "-f", string(ApplicationKF)})
	return
}

func (e2e *HAOnMinio) verifyApplicationIsRunning() (err error) {
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

func (e2e *HAOnMinio) verifyAllVolumeReplicasAreRunning() (err error) {
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

func (e2e *HAOnMinio) verifyMinioApplicationIsLaunchedSuccessfullyOnVolume() (err error) {
	err = e2e.verifyApplicationIsRunning()
	if err != nil {
		return
	}

	// check if volume is running
	return e2e.verifyAllVolumeReplicasAreRunning()
}

func (e2e *HAOnMinio) verifyPVCIsBound() (err error) {
	if e2e.appVerifier == nil {
		err = fmt.Errorf("nil application verifier: possible error '%s'", e2e.errors[ApplicationVerifyFileEI])
		return
	}

	// is condition satisfied
	_, err = e2e.appVerifier.IsCondition(PVCAlias, verify.PVCBoundCond)
	return
}

func (e2e *HAOnMinio) verifyPVIsDeployed() (err error) {
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

func (e2e *HAOnMinio) minioApplicationIsLaunchedSuccessfullyOnVolume() (err error) {
	return e2e.verifyMinioApplicationIsLaunchedSuccessfullyOnVolume()
}

func (e2e *HAOnMinio) launchMinioClientPutJob() (err error) {
	_, err = kubectl.New().Run([]string{"apply", "-f", string(AppClientPutKF)})
	return
}

func (e2e *HAOnMinio) cordonTheNodeThatHostsTheMinioPod() (err error) {
	if e2e.appVerifier == nil {
		err = fmt.Errorf("nil application verifier: possible error '%s'", e2e.errors[ApplicationVerifyFileEI])
		return
	}

	// is action satisfied
	_, err = e2e.appVerifier.IsAction(AppPodAlias, verify.CordonNodeWithOldestPodAction)
	return
}

func (e2e *HAOnMinio) deleteThisMinioPod() (err error) {
	if e2e.appVerifier == nil {
		err = fmt.Errorf("nil application verifier: possible error '%s'", e2e.errors[ApplicationVerifyFileEI])
		return
	}

	// is action satisfied
	_, err = e2e.appVerifier.IsAction(AppPodAlias, verify.DeleteOldestPodAction)
	return
}

func (e2e *HAOnMinio) verifyMinioIsRedeployedSuccessfully() (err error) {
	return e2e.verifyApplicationIsRunning()
}

func (e2e *HAOnMinio) launchMinioClientGetJob() (err error) {
	_, err = kubectl.New().Run([]string{"apply", "-f", string(AppClientGetKF)})
	return
}

func (e2e *HAOnMinio) deployMinioClientConfigSetWithMinioServerIP() (err error) {
	f, err := fetch.NewKubeResourceFetch(ApplicationIF)
	if err != nil {
		return
	}

	// fetch service ip
	data, err := f.Fetch(AppServiceAlias, fetch.ServiceIPProperty)
	if err != nil {
		return
	}

	// service ip
	ip := data[0]

	templateContent, err := exec.NewShellExec("cat").Execute([]string{string(AppClientConfigTKF)})
	if err != nil {
		return
	}

	// parse template file
	t, err := template.New("AppClientConfigTKF").Parse(templateContent)
	if err != nil {
		return
	}

	// render template file with values
	var output bytes.Buffer
	values := map[string]string{
		"ip": ip,
	}
	err = t.Execute(&output, values)
	if err != nil {
		return
	}

	// kubectl apply this rendered file
	err = kubectl.ApplyStdIn(output.Bytes())
	if err != nil {
		return
	}

	// verify app client config deployment
	appcConfig, err := verify.NewKubeInstallVerify(AppClientConfigIF)
	if err != nil {
		return
	}

	// is application client config deployed
	_, err = appcConfig.IsDeployed()
	return
}

func (e2e *HAOnMinio) verifyDataIsPutToMinioServer() (err error) {
	// has put job completed successfully ?
	v, err := verify.NewKubeInstallVerify(AppClientJobIF)
	if err != nil {
		return
	}

	_, err = v.IsCondition(PutJobAlias, verify.JobCompletedCond)
	return
}

func (e2e *HAOnMinio) verifyDataIsAvailableAtMinioServer() (err error) {
	// has get job completed successfully ?
	v, err := verify.NewKubeInstallVerify(AppClientJobIF)
	if err != nil {
		return
	}

	_, err = v.IsCondition(GetJobAlias, verify.JobCompletedCond)
	return
}

func FeatureContext(s *godog.Suite) {
	e2e := &HAOnMinio{
		errors: map[errorIdentity]error{},
	}

	// before feature run
	s.BeforeFeature(e2e.withOperatorVerifier)
	s.BeforeFeature(e2e.withApplicationVerifier)
	s.BeforeFeature(e2e.withVolumeVerifier)

	// after feature run
	s.AfterFeature(e2e.tearDown)

	s.Step(`^I have a kubernetes multi node cluster$`, e2e.iHaveAKubernetesMultiNodeCluster)
	s.Step(`^this cluster has volume operator installed$`, e2e.thisClusterHasVolumeOperatorInstalled)
	s.Step(`^I launch minio application on volume$`, e2e.iLaunchMinioApplicationOnVolume)
	s.Step(`^wait for "([^"]*)"$`, e2e.waitFor)
	s.Step(`^verify minio application is launched successfully on volume$`, e2e.verifyMinioApplicationIsLaunchedSuccessfullyOnVolume)
	s.Step(`^verify PVC is bound$`, e2e.verifyPVCIsBound)
	s.Step(`^verify PV is deployed$`, e2e.verifyPVIsDeployed)
	s.Step(`^minio application is launched successfully on volume$`, e2e.minioApplicationIsLaunchedSuccessfullyOnVolume)
	s.Step(`^launch minio client put job$`, e2e.launchMinioClientPutJob)
	s.Step(`^cordon the node that hosts the minio pod$`, e2e.cordonTheNodeThatHostsTheMinioPod)
	s.Step(`^delete this minio pod$`, e2e.deleteThisMinioPod)
	s.Step(`^verify minio is redeployed successfully$`, e2e.verifyMinioIsRedeployedSuccessfully)
	s.Step(`^launch minio client get job$`, e2e.launchMinioClientGetJob)
	s.Step(`^deploy minio client config set with minio server IP$`, e2e.deployMinioClientConfigSetWithMinioServerIP)
	s.Step(`^verify data is put to minio server$`, e2e.verifyDataIsPutToMinioServer)
	s.Step(`^verify data is available at minio server$`, e2e.verifyDataIsAvailableAtMinioServer)
}
