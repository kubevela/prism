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
	"context"
	"fmt"
	"strings"

	clustergatewayv1alpha1 "github.com/oam-dev/cluster-gateway/pkg/apis/cluster/v1alpha1"
	clustergatewaycommon "github.com/oam-dev/cluster-gateway/pkg/common"
	clustergatewayconfig "github.com/oam-dev/cluster-gateway/pkg/config"
	corev1 "k8s.io/api/core/v1"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ocmclusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubevela/prism/pkg/util/apiserver"
	"github.com/kubevela/prism/pkg/util/singleton"
)

// Get finds a resource in the storage by name and returns it.
func (in *Cluster) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return NewClusterClient(singleton.GetKubeClient()).Get(ctx, name)
}

// List selects resources in the storage which match to the selector. 'options' can be nil.
func (in *Cluster) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	return NewClusterClient(singleton.GetKubeClient()).List(ctx, apiserver.NewMatchingLabelSelectorFromInternalVersionListOptions(options))
}

func extractLabels(labels map[string]string) map[string]string {
	_labels := make(map[string]string)
	for k, v := range labels {
		if !strings.HasPrefix(k, clustergatewayconfig.MetaApiGroupName) {
			_labels[k] = v
		}
	}
	return _labels
}

func newCluster(obj client.Object) *Cluster {
	cluster := &Cluster{}
	cluster.SetGroupVersionKind(ClusterGroupVersionKind)
	if obj != nil {
		cluster.SetName(obj.GetName())
		cluster.SetCreationTimestamp(obj.GetCreationTimestamp())
		cluster.SetLabels(extractLabels(obj.GetLabels()))
		if annotations := obj.GetAnnotations(); annotations != nil {
			cluster.Spec.Alias = annotations[AnnotationClusterAlias]
		}
	}
	cluster.Spec.Accepted = true
	cluster.Spec.Endpoint = ClusterBlankEndpoint
	metav1.SetMetaDataLabel(&cluster.ObjectMeta, LabelClusterControlPlane, fmt.Sprintf("%t", obj == nil))
	return cluster
}

// NewLocalCluster return the local cluster
func NewLocalCluster() *Cluster {
	cluster := newCluster(nil)
	cluster.SetName(ClusterLocalName)
	cluster.Spec.CredentialType = CredentialTypeInternal
	return cluster
}

// NewClusterFromSecret extract cluster from cluster secret
func NewClusterFromSecret(secret *corev1.Secret) (*Cluster, error) {
	cluster := newCluster(secret)
	cluster.Spec.Endpoint = string(secret.Data["endpoint"])
	if metav1.HasLabel(secret.ObjectMeta, clustergatewaycommon.LabelKeyClusterEndpointType) {
		cluster.Spec.Endpoint = secret.GetLabels()[clustergatewaycommon.LabelKeyClusterEndpointType]
	}
	if cluster.Spec.Endpoint == "" {
		return nil, NewEmptyEndpointClusterSecretError()
	}
	if !metav1.HasLabel(secret.ObjectMeta, clustergatewaycommon.LabelKeyClusterCredentialType) {
		return nil, NewEmptyCredentialTypeClusterSecretError()
	}
	cluster.Spec.CredentialType = clustergatewayv1alpha1.CredentialType(
		secret.GetLabels()[clustergatewaycommon.LabelKeyClusterCredentialType])
	return cluster, nil
}

// NewClusterFromManagedCluster extract cluster from ocm managed cluster
func NewClusterFromManagedCluster(managedCluster *ocmclusterv1.ManagedCluster) (*Cluster, error) {
	if len(managedCluster.Spec.ManagedClusterClientConfigs) == 0 {
		return nil, NewInvalidManagedClusterError()
	}
	cluster := newCluster(managedCluster)
	cluster.Spec.Accepted = managedCluster.Spec.HubAcceptsClient
	cluster.Spec.CredentialType = CredentialTypeOCMManagedCluster
	return cluster, nil
}
