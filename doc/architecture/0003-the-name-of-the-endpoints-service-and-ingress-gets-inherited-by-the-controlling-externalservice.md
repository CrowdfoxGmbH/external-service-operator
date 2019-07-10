# 3. The name of the Endpoints, Service and Ingress gets inherited by the controlling ExternalService

Date: 2019-07-17

## Status

Accepted

## Context

To easify finding the according Endpoints, Services and Ingress Ressources, they are named exactly the same as the Externalservice Ressource.

Nethertheless, of course the Owner will be set correctly as well as every ressource gets the label:
```
app= <external-servicename>
```

## Decision

The change that we're proposing or have agreed to implement.

## Consequences

In order to break this scheme, we have to adjust all the places where resources are fetched and applied.
