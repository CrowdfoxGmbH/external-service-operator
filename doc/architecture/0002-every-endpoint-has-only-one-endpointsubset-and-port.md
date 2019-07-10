# 2. Every Endpoint has only one EndpointSubset and Port

Date: 2019-07-17

## Status

Accepted

## Context

The ExternalService Operator uses a Custom Resource Definition [../../deploy/crds/eso_v1alpha1_externalservice_cr.yaml]() in Order to describe external services which made available for the Kubernetes Cluster as they would be running inside the cluster.

This makes it possible to expose services to a central in the cluster running ingress container.

For both mechanisms, but especially the ingress routing one, it is not needed to have more than one EndpointSubset or having more than one port.

As you might see, the Ingress Ressource Rules are only capable of having one Service and one Port, so this are the Kubernetes Resources managed by the Operator
```
apiVersion: v1
items:
- apiVersion: extensions/v1beta1
  kind: Ingress
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"eso.crowdfox.com/v1alpha1","kind":"ExternalService","metadata":{"annotations":{"nginx.ingress.kubernetes.io/rewrite-target":"/"},"name":"example-externalservice","namespace":"default"},"spec":{"hosts":[{"host":"subdomain.example.com","path":"/"}],"ips":["192.168.0.10","192.168.0.11","192.168.0.12"],"port":80}}
      nginx.ingress.kubernetes.io/rewrite-target: /
    creationTimestamp: 2019-07-16T13:32:13Z
    generation: 4
    labels:
      app: example-externalservice
      serviceType: external
    name: example-externalservice
    namespace: default
    resourceVersion: "39534"
    selfLink: /apis/extensions/v1beta1/namespaces/default/ingresses/example-externalservice
    uid: 1f255906-a7ce-11e9-8a65-0800278d3f0c
  spec:
    rules:
    - host: test.crowdfox.test
      http:
        paths:
        - backend:
            serviceName: example-externalservice
            servicePort: 9090
          path: /
  status:
    loadBalancer: {}
- apiVersion: v1
  kind: Service
  metadata:
    creationTimestamp: 2019-07-16T13:32:13Z
    labels:
      app: example-externalservice
      serviceType: external
    name: example-externalservice
    namespace: default
    resourceVersion: "39533"
    selfLink: /api/v1/namespaces/default/services/example-externalservice
    uid: 1f14497e-a7ce-11e9-8a65-0800278d3f0c
  spec:
    clusterIP: None
    ports:
    - port: 9090
      protocol: TCP
      targetPort: 9090
    sessionAffinity: None
    type: ClusterIP
  status:
    loadBalancer: {}
- apiVersion: v1
  kind: Endpoints
  metadata:
    creationTimestamp: 2019-07-16T13:32:13Z
    labels:
      app: example-externalservice
      serviceType: external
    name: example-externalservice
    namespace: default
    ownerReferences:
    - apiVersion: eso.crowdfox.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: ExternalService
      name: example-externalservice
      uid: 1eeffc8f-a7ce-11e9-8a65-0800278d3f0c
    resourceVersion: "42925"
    selfLink: /api/v1/namespaces/default/endpoints/example-externalservice
    uid: 1f041bf1-a7ce-11e9-8a65-0800278d3f0c
  subsets:
  - notReadyAddresses:
    - ip: 192.168.0.10
    - ip: 192.168.0.12
    ports:
    - port: 9090
      protocol: TCP
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```


## Decision

The implementation relies on this decision, that there is always exactly one EndpointSubset and in that EndpointSubset exactly one Port defined.

## Consequences

When this wants to be extended, every place where Endpoints are touched or created has to be refactored. Especially the UpdateEndpoint behaviour and the Healthchecks
