# Prism
[![Go Report Card](https://goreportcard.com/badge/github.com/kubevela/kubevela)](https://goreportcard.com/report/github.com/kubevela/kubevela)
[![codecov](https://codecov.io/gh/kubevela/prism/branch/master/graph/badge.svg)](https://codecov.io/gh/kubevela/vela-prism)
[![LICENSE](https://img.shields.io/github/license/kubevela/prism.svg?style=flat-square)](/LICENSE)

## Introduction

**Prism** provides API Extensions to the core [KubeVela](https://github.com/kubevela/kubevela). 

### apiserver

The vela-prism is an apiserver which leverages the Kubernetes Aggregated API capability to provide native interface for users.

#### ApplicationResourceTracker

The original ResourceTracker in KubeVela is one kind of cluster-scoped resource (for some history reasons), which makes it hard for cluster administrator to assign privilege.
The ApplicationResourceTracker is a kind of namespace-scoped resource, which works as a delegator to the original ResourceTracker.
It does not need extra storages but can project requests to ApplicationResourceTracker to underlying ResourceTrackers.
Therefore, it is possible for cluster administrator to assign ApplicationResourceTracker permissions to users.

After installing vela-prism in your cluster, you can run `kubectl get apprt` to view ResourceTrackers.
