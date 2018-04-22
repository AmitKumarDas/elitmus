## Litmus
Litmus test your application without learning curves

### Motivation
>Testing as they say "can show presence of bugs, and not their absence". 

However, we can try to eliminate bugs if we are able to let the stakeholders participate in probing for evidences of bugs. Each stakeholder, be it the developer or the analyst or the product manager or the end user or the tester & so on bring their own set of scenarios that can have a snowball effect in finding bugs that might be hidden deep underneath.

### Beliefs
- Testing similar to its code, improves with the community participation. 
- Litmus keeps end user in mind while designing its test scenarios.

## Development

### Compile
- Make use of Makefile
- Run below command to compile the code
 - `make`

### Build & Push the Docker image
- `sudo docker build . -t openebs/litmus:latest`
- `sudo docker push openebs/litmus:latest`

## Run

### Run a specific testcase as a Kubernetes Job
- Assumption: Provider specific K8s resources should be deployed
  - NOTE: This is a one time activity
  - e.g. this installs openebs operator
  - `kubectl apply -f tests/openebs/openebs-operator-v0.5.3.yaml`
  - `kubectl apply -f tests/openebs/openebs-storage-classes-v0.5.3.yaml`
- Install RBAC policies required for K8s job to execute
  - NOTE: This is a one time activity
  - `kubectl apply -f rbac.yaml`
- Below commands runs the testcase **mysql_resiliency_with_3_reps** as a Kubernetes Job
- `kubectl -n litmus create configmap omrwtr-application-launch --from-file=config=tests/openebs/mysql_resiliency_with_3_reps/application-launch.yaml`
- `kubectl apply -f tests/openebs/mysql_resiliency_with_3_reps/test-the-feature.yaml`

## Troubleshooting

### Check the pod logs
- `kubectl get pod -a`
- `kubectl logs <recent_pod_that_errored_out>`

### Analyze via docker run
- Try running the testcase via docker run to eliminate Dockerfile related issues
  - Below command troubleshoots the testcase **mysql_resiliency_with_3_reps**
  - `sudo docker run -w /go/src/github.com/AmitKumarDas/litmus/cmd/mysql_resiliency_with_3_reps -it openebs/litmus:latest godog e2e.feature`

## Appendix
- Scenario: An example of the system's behavior from one or more user's perspectives
