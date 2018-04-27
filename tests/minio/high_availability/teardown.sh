#!/usr/bin/env sh

set -o nounset

CURDIR=`pwd`

############################################
# delete the test
############################################

# delete the kubernetes job that is responsible to run the test into completion
kubectl delete -f ./test-this-feature.yaml

############################################
# delete the associated configmaps
############################################

# delete the test feature related verify configs
kubectl delete -f ./test-this-feature-configs.yaml
# delete application client related configs
kubectl -n litmus delete configmap ha-minio-app-client-configs
# delete application client's get job config
kubectl -n litmus delete configmap ha-minio-app-client-get-job
# delete application client's put job config
kubectl -n litmus delete configmap ha-minio-app-client-put-job
# delete application launch config
kubectl -n litmus delete configmap ha-minio-app-launch

# delete the application & related resources incase the job was restarted 
# due to a failure
# NOTE:
#   When a job completes successfully, the job itself will clear these resources
# In other words, following cmd will result in failure when jobs are ran 
# successfully.
kubectl delete -f ./application-launch.yaml

cd ${CURDIR}
