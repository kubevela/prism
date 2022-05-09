/*
Copyright 2022 The KubeVela Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/oam-dev/cluster-gateway/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/pflag"
)

const (
	// ClusterLocalName name for the hub cluster
	ClusterLocalName = "local"
	// CredentialTypeInternal identifies the cluster from internal kubevela system
	CredentialTypeInternal v1alpha1.CredentialType = "Internal"
	// CredentialTypeOCMManagedCluster identifies the ocm cluster
	CredentialTypeOCMManagedCluster v1alpha1.CredentialType = "ManagedCluster"
	// ClusterBlankEndpoint identifies the endpoint of a cluster as blank (not available)
	ClusterBlankEndpoint = "-"
)

// StorageNamespace refers to the namespace of cluster secret, usually same as the core kubevela system namespace
var StorageNamespace string

// AddClusterFlags add flags for cluster api
func AddClusterFlags(set *pflag.FlagSet) {
	set.StringVarP(&StorageNamespace, "storage-namespace", "", "vela-system",
		"The namespace that stores cluster secrets or OCM ManagedClusters.")
}
