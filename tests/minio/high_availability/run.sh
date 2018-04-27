#!/usr/bin/env sh

set -o errexit
set -o nounset

CURDIR=`pwd`

#####################################################
# pre-requisites before test run
#####################################################

# create verify configmaps
kubectl apply -f ./test-this-feature-configs.yaml
# create minio client related configmaps
kubectl -n litmus create configmap ha-minio-app-client-configs --from-file=config=application-client-configs.yaml
# embed minio client's put job in a configmap
kubectl -n litmus create configmap ha-minio-app-client-put-job --from-file=config=application-client-put-job.yaml
# embed minio client's get job in a configmap
kubectl -n litmus create configmap ha-minio-app-client-get-job --from-file=config=application-client-get-job.yaml
# embed minio application in a configmap
kubectl -n litmus create configmap ha-minio-app-launch --from-file=config=application-launch.yaml

####################################################
# run the test
####################################################

# launch the kubernetes job responsible to run the test into completion
kubectl apply -f test-this-feature.yaml

cd ${CURDIR}
