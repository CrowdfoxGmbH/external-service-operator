# 6. Only one Probe of HTTP, TCP is possible

Date: 2021-03-22

## Status

Accepted

## Context

The Kubernetes Probe Resource Structure would allow adding multiple types of Healthchecks at once. This would be:
* exec
* httpGet
* tcpSocket

However currently v1.18 v1.20 when one tries to add more than one Type, like:
```
    readinessProbe:
      exec:
        command:
        - "/usr/bin/sh"
        - "-c"
        - "echo Hello World"
      httpGet:
        host: localhost
        port: 80
      tcpSocket:
        host: localhost
        port: 80
```

kubernetes will not validate the Resource with following message:

```
# pods "test-probes" was not valid:
# * spec.containers[0].readinessProbe.httpGet: Forbidden: may not specify more than 1 handler type
# * spec.containers[0].readinessProbe.tcpSocket: Forbidden: may not specify more than 1 handler type
```


## Decision

It is decided to go with the same logic and ensure that only one type will be accepted, so the Operator is not supposed to handle the case more than one Types are given.
Further it follows the same logic and will use and check HTTP first than TCP last. Exec healthchecks don't make sense and will be ignored completly

## Consequences

If the behaviour of kubernetes changes this might also have to change but this is very unlikely.
More likely is, even when kubernetes changes this behaviour, that the service operator will stick to the described behaviour.
