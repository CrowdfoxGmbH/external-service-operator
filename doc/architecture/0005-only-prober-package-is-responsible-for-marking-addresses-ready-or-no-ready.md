# 5. Only Prober Package is responsible for markind Addresses ready or no ready

Date: 2019-07-24

## Status

Accepted

## Context

As the Endpoint ressource has two controlling structures.
1. Endpoint Reconciler
2. Prober (ProbeManager and according Workers)



## Decision

There is the need to seperate concerns in that resource. Therefore:

### Endpoint Reconciler

Is responsible for modifying:
* ObjectMeta
* General Structure for Subsets
* It makes sure exactly the same IPs in the Externalservice ressource exist also in the Endpoint
* If not it will add the missing IPs always to the NotReadyAddresses List

### Prober

The Prober package is soley responsible for moving Addresses from Addresses to NotReadyAddresses and vice versa.


## Consequences

As a Consequence the EndpointReconciler conciders Endpoints to be equal no matter if IPs are in Ready or NotReady state. (Means it will merge both lists and compare that list as they would have been sets)
No component except the prober package is allowed to move IPs between **ready** and **not ready**
