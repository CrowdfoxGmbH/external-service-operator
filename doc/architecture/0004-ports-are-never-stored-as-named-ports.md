# 4. Ports are never stored as named Ports (strings)

Date: 2019-07-18

## Status

Accepted

## Context

Actually it does not make sense to store ports as strings, as those Backends are not in the cluster so they don't have names and can only be referenced by a port number. Nevertheless the Kuberenetes API makes it possible to store ports as string which will not be used at the endpoints generated by the External Service Operator.

## Decision

The change that we're proposing or have agreed to implement.

## Consequences

Ports on Endpoints must be always integers. Also in the ExternalServiceSpec