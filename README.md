Upgrading kubernetes deployments without downtime
=================================================

This repo contains an example deployment to demonstrate edge cases that can cause HTTP requests to fail when performing a rolling upgrade of a kubernetes deployment.


## Example App

The example app consists of a deployment that is exposed by an ingress resource. It is assumed the nginx-ingress controller is used to expose the application.


```
[client] ---> [nginx ingress controller]  ---> [Pod IPs]
```
### Running the tests

To test the resiliency of the deployment against failures during upgrades perform the following procedure:

* Edit `resources/deployment.yaml` and bump the `VERSION` environment variable to simulate a version bump
* Run `make load` in a terminal window to start sending requests to the application
* While the load test is running in the shell run `make deploy` to trigger the deployment
* Check the load tests for any errors

NOTE: The load tests only performs `POST` requests on purpose to not have the ingress controller (e.g. nginx) automatically retry idempotent requests hiding the underlying problem.

## Edge cases
### Premature activation of new endpoints
By default when a pod enters the running state it is considered `Ready` and the Pod ip is added as an active endpoint to the `Endpoints` object of the backing service.
If the app is not yet fully running it might receive traffic to early causing those requests to fail.

This edge case can be prevented by adding a [ReadinessProbe](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#when-should-you-use-a-readiness-probe) to the container.

Fix: `kustomize edit add patch patches/readiness-probe.yaml`

### Premature shutdown of old endpoints
When old Pods are deleted the are send the `TERM` signal and after the grace period they are killed.
If the application terminates when receiving the `TERM` signal it might drop (longer running) in-flight requests.

To avoid this edge case the application needs to perform a *graceful* shutdown draining the local request queue and only shutting down after all pending requests have been completed.
The timeout for a graceful shutdown needs to be selected to according the maximum time a request is allowed to take.

Fix: `kustomize edit add patch patches/graceful-shutdown.yaml` 


### Stale ingress upstream endpoints

When an old Pod is deleted the Pod enters the `Terminating` state it is immediately removed from the list of active endpoints. But there might be a considerable delay before this change is reflected in the configuration of the ingress controller (nginx). This means that even after the endpoint is removed from the list it still might receive new traffic from the ingress controller that still has the endpoint configured as an upstream.
Typically this edge case is not covered by the graceful shutdown procedure of an application as the listener is closed at the very start of the shutdown and the application only cares about exiting connections/requests.

To fix this the shutdown of the applications listener socket needs to be delayed for terminating pods to allow the configuration in the ingress controller to catch up. As most application don't support this a [PreStop Hook](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks) can be used to delay sending the `TERM` signal to the container which in turn triggers the graceful shutdown of the backend server.

Fix: `kustomize edit add patch patches/pre-stop-hook.yaml`


## Summary

To prevent loosing requests when performing a deployment the following things needs to be considered:

* Readiness Probe that siganls when the container is ready to receive traffic
* Graceful shutdown on `TERM` signal that delays process exit until in-flight requests are completed. Timeout needs to be `> max request time`
* `prepStop` hook that waits for ~ 15 seconds before shutting down the listener socket
* Pods `terminationGracePeriodSeconds` needs to be `> max request time + pre stop time (15 seconds)`  