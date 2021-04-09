External Service Operator
============================

This project was Bootstrapped with the [operator-sdk](https://github.com/operator-framework/operator-sdk) together with this tutorial: https://medium.com/faun/writing-your-first-kubernetes-operator-8f3df4453234

The External Service Operator is meant to manage Services which are outside of the Kubernetes Cluster but should be used "Cloud Native" inside the cluster.
The Operator has following features:
* Creates Endpoints, Services and Ingresses for an external Service for a given list of (IP, Port) tuples.
* It is possible to set custom ingress annotations
* Is doing healthchecks and remove IPs from Endpoints when they fail.

You can find more details in the CRD descriptions.

Installation
--------------

### Helm

The easiest way to install the External Service Operator is via helm:
https://github.com/CrowdfoxGmbH/cfcharts/tree/master/charts/external-service-operator

### Manually applying RDs

Therefore `kubectl` apply:
* RBACs (only needed when you have RBAC enabled, which you absolutly should!)
  * deploy/namespace.yaml
  * deploy/clusterrole.yaml
  * deploy/service\_account.yaml
  * deploy/role\_binding.yaml
* Custom Resource Definitions (CRDS)
  * deploy/crds/eso\_v1alpha1\_externalservice\_crd.yaml
* Operator
  * deploy/operator.yaml *you may want to remove the development option* `--zap-devel`

Then if you like, you can deploy an externalservice like: `deploy/crds/eso\_v1alpha1\_externalservice\_crd.yaml`

ExternalService Resource Description
------------------------------------

Together with the Operator a new CRD will be deployed. Best way to figure out Options is to look into:
[./pkg/apis/eso/v1alpha1/externalservice\_types.go]()

A sample CRD can be found in [./deploy/crds/eso\_v1alpha1\_externalservice\_crd.yaml]()

Note, that you can use only [HTTPGetActions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#httpgetaction-v1-core) and [TCPSocketActions](https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#tcpsocketaction-v1-core) Probes. ExecAction and TCPSocketAction may follow later.

A very complex example of an External Service could look like:

```YAML
apiVersion: eso.crowdfox.com/v1alpha1
kind: ExternalService
metadata:
  annotations:
    # those annotations will be added to the ingress ressource
    traefik.ingress.kubernetes.io/preserve-host: "true"
  name: complex-example
  namespace: external-service
spec:
  hosts:
  - host: static1.mydomain.com
    path: ""
  - host: static2.mydomain.com
    path: ""
  - host: mydomain.com
    path: ""
  - host: www.mydomain.com
    path: ""
  ips:
  - 10.0.100.10
  - 10.0.100.11
  port: 80
  readinessProbe:
    failureThreshold: 3
    httpGet:
      host: www.mydomain.com
      httpHeaders:
      - name: X-Forwarded-Proto
        value: https
      path: /healthcheck
      port: 8080
      scheme: HTTP
    periodSeconds: 10
    successThreshold: 2
    timeoutSeconds: 2
```

Development
-----------

Before you start, you may or better should read architectural decisions. You can find them in in [./doc/architecture]()

### Setup Environment

* Install Golang: https://golang.org/doc/install
* Install Operator SDK: https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md
* Set env: `export GO111MODULE=on`
* Install Dependencies locally: `go mod vendor`
* Run `go test github.com/CrowdfoxGmbH/external-service-operator/pkg/controller/externalservice github.com/CrowdfoxGmbH/external-service-operator/pkg/prober`

### Build Docker container

* operator-sdk build crowdfox/external-service-operator:<tag>
* docker login (with user crowdfox and password from OnePassword)
* docker push crowdfox/external-service-operator:<tag>
* change image reference: `sed -i 's|image: .*|image: crowdfox/external-service-operator:<tag>|g' deploy/operator.yaml`
* minikube start
* apply Ressources from deployment folder

### Exploration Tests

* When the operator is running, you can apply the crds.
* Then you can spin up an HTTP Server like: `docker run --rm -it --name my-apache-app -p 9090:80 -v /var/www/html:/usr/local/apache2/htdocs/ httpd:2.4`
* Check the endpoint: `kubectl get endpoints example-externalservice -o yaml | grep -A 10 subsets`. The IP Address should be in NotReadyAddresses.
* Change the externalservice resource to the IP (your IP) and port of the HTTP Container.
* Wait some time (depending on the settings for the probes). Then the IP in the endpoint should be in Addresses

### Readiness Probe

The implementation is highly inspired by (not to say copied from) the kubelet prober implementation:
* [probeManager](https://github.com/kubernetes/kubernetes/blob/8f41397210e03a328b684a042b96dfcdca066fd5/pkg/kubelet/prober/.)
* [probeWorker](https://github.com/kubernetes/kubernetes/blob/8f41397210e03a328b684a042b96dfcdca066fd5/pkg/kubelet/prober/worker.go)

Therefore the Reconciler ist starting or stopping probeWorker and modifies the Endpoint CRs when the state of the backend instance changed.

### Helpful Links

1. of course the operator-sdk documentation in general but in special: [https://github.com/operator-framework/operator-sdk/blob/master/doc/user/unit-testing.md]()
1. Go API for Kubernetes: [https://godoc.org/k8s.io/]() where you can find type Intefaces
1. Kubernetes API in general: [https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11]()

