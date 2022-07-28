# Prism
[![UnitTest](https://github.com/kubevela/prism/actions/workflows/unit-test.yml/badge.svg)](https://github.com/kubevela/prism/actions/workflows/unit-test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubevela/prism)](https://goreportcard.com/report/github.com/kubevela/prism)
[![codecov](https://codecov.io/gh/kubevela/prism/branch/master/graph/badge.svg)](https://codecov.io/gh/kubevela/vela-prism)
[![LICENSE](https://img.shields.io/github/license/kubevela/prism.svg?style=flat-square)](/LICENSE)
[![Total alerts](https://img.shields.io/lgtm/alerts/g/kubevela/prism.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/kubevela/prism/alerts/)

## Introduction

**Prism** provides API Extensions to the core [KubeVela](https://github.com/kubevela/kubevela).
It works as a Kubernetes [Aggregated API Server](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/apiserver-aggregation/).

![PrismArch](https://github.com/kubevela/prism/blob/master/hack/prism-arch.jpg)

## Installation

### From chart repository

Add the chart repository
```shell
helm repo add prism https://charts.kubevela.net/prism
helm repo update
```

Install helm chart
```shell
helm install vela-prism prism/vela-prism -n vela-system
```

### From this repo
Clone this repo and run `helm install vela-prism charts/ --namespace vela-system`.

## Modules

### apiserver

The vela-prism is an apiserver which leverages the Kubernetes Aggregated API capability to provide native interface for users.

#### ApplicationResourceTracker

The original ResourceTracker in KubeVela is one kind of cluster-scoped resource (for some history reasons), which makes it hard for cluster administrator to assign privilege.
The ApplicationResourceTracker is a kind of namespace-scoped resource, which works as a delegator to the original ResourceTracker.
It does not need extra storages but can project requests to ApplicationResourceTracker to underlying ResourceTrackers.
Therefore, it is possible for cluster administrator to assign ApplicationResourceTracker permissions to users.

After installing vela-prism in your cluster, you can run `kubectl get apprt` to view ResourceTrackers.

#### Cluster

In vela-prism, Cluster API is also introduced which works as a delegator to the ClusterGateway object.
The original ClusterGateway object contains the credential information.
This makes the exposure of ClusterGateway access can be dangerous.
The Cluster object provided in prism, on the other hand, only expose metadata of clusters to accessor.
Therefore, the credential information will be secured and the user can also use the API to access the cluster list.

After installing vela-prism in your cluster, you can run `kubectl get vela-clusters` to view all the installed clusters.

> Notice that the vela-prism bootstrap parameter contains `--storage-namespace`, which identifies the underlying namespace for storing cluster secrets and the OCM managed cluster.

### Grafana related APIs

![PrismGrafanaArch](https://github.com/kubevela/prism/blob/master/hack/prism-grafana-arch.jpg)

#### Grafana

In vela-prism, you can store grafana access into Grafana object. The Grafana object is projected into secrets in `o11y-system` (this can be configured through `--observability-namespace` parameter).
The secret embeds the access endpoint and credential (either username/password for BasicAuth or token for BearerToken) for grafana. These will be used for communicating with Grafana APIs. Example of Grafana object is shown below.

```yaml
apiVersion: o11y.prism.oam.dev/v1alpha1
kind: Grafana
metadata:
  name: example
spec:
  access:
    username: admin
    password: kubevela
  endpoint: https://grafana.o11y-system:3000/
```

#### GrafanaDashboard & GrafanaDatasource

After creating the Grafana object into the control plane, you are now able to manipulate Grafana resources through Kubernetes APIs now. 
Currently, vela-prism provides proxies for GrafanaDatasource and GrafanaDashboard.
Their names are constructed by two parts, its original name and the backend grafana name. 

For example, you can create a new GrafanaDashboard by applying the following YAML file. This will use the Grafana object above as the access credential, and call the Grafana APIs to create a new dashboard.

You can also update or delete dashboards or datasources. The spec part of GrafanaDashboard and GrafanaDatasource are directly projected into API request body.

```yaml
apiVersion: o11y.prism.oam.dev/v1alpha1
kind: GrafanaDashboard
metadata:
  name: dashboard-test@example
spec:
  title: New dashboard
  tags: []
  style: dark
  timezone: browser
  editable: true
  hideControls: false
  graphTooltip: 1
  panels: []
  time:
    from: now-6h
    to: now
  timepicker:
    time_options: []
    refresh_intervals: []
  templating:
    list: []
  annotations:
    list: []
  refresh: 5s
  schemaVersion: 17
  version: 0
  links: []
```

Another example for GrafanaDatasource.

```yaml
apiVersion: o11y.prism.oam.dev/v1alpha1
kind: GrafanaDatasource
metadata:
  name: prom-test@example
spec:
  access: proxy
  basicAuth: false
  isDefault: false
  name: ExamplePrometheus
  readOnly: true
  type: prometheus
  url: https://prometheus-server.o11y-system:9090
```

#### verse operator pattern

To operate Grafana instances in Kubernetes, there are also [Grafana operators](https://github.com/grafana-operator/grafana-operator) to help manage Grafana configurations.

Compared to operator pattern, the aggregator pattern has pros and cons.

##### Pros
- There is no data consistency problem between the CustomResource and the Grafana underlying storage.
- API requests are made in time. The response is also immediate. No need for checking CustomResource status repeatedly.
- No reconciles. No need to hold CPUs and memories as controller does.
- Easy to connect with third-party Grafana instance. For example, Grafana from cloud providers.

##### Cons
- Cannot persist data outside Grafana storage. Once grafana is broken, the CustomResource will be unavailable as well.

The main drawback for vela-prism compared to operator pattern is that it cannot persist configurations. However, this can be solved through using KubeVela application to manage those configurations.
For example, you can write KubeVela applications to hold the dashboard configurations, instead of create another separate CustomResource. With KubeVela application, leverage GitOps is also possible.