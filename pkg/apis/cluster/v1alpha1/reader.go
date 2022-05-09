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
	"strings"

	clustergatewayv1alpha1 "github.com/oam-dev/cluster-gateway/pkg/apis/cluster/v1alpha1"
	clustergatewaycommon "github.com/oam-dev/cluster-gateway/pkg/common"
	clustergatewayconfig "github.com/oam-dev/cluster-gateway/pkg/config"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	apitypes "k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ocmclusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubevela/prism/pkg/util/singleton"
)

// Get finds a resource in the storage by name and returns it.
func (in *Cluster) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	if name == ClusterLocalName {
		return NewLocalCluster(), nil
	}
	key := apitypes.NamespacedName{Name: name, Namespace: StorageNamespace}
	var cluster *Cluster
	secret := &corev1.Secret{}
	err := singleton.GetKubeClient().Get(ctx, key, secret)
	var secretErr error
	if err == nil {
		if cluster, secretErr = NewClusterFromSecret(secret); secretErr == nil {
			return cluster, nil
		}
	}
	if err != nil && !apierrors.IsNotFound(err) {
		secretErr = err
	}

	managedCluster := &ocmclusterv1.ManagedCluster{}
	err = singleton.GetKubeClient().Get(ctx, key, managedCluster)
	var managedClusterErr error
	if err == nil {
		if cluster, managedClusterErr = NewClusterFromManagedCluster(managedCluster); managedClusterErr == nil {
			return cluster, nil
		}
	}

	if err != nil && !apierrors.IsNotFound(err) && !meta.IsNoMatchError(err) {
		managedClusterErr = err
	}

	errs := utilerrors.NewAggregate([]error{secretErr, managedClusterErr})
	if errs == nil {
		return nil, apierrors.NewNotFound(ClusterGroupResource, name)
	} else if len(errs.Errors()) == 1 {
		return nil, errs.Errors()[0]
	} else {
		return nil, errs
	}
}

// List selects resources in the storage which match to the selector. 'options' can be nil.
func (in *Cluster) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	clusters := &ClusterList{}
	if options == nil || options.LabelSelector == nil || options.LabelSelector.Empty() {
		local := NewLocalCluster()
		clusters.Items = []Cluster{*local}
	}

	secrets := &corev1.SecretList{}
	err := singleton.GetKubeClient().List(ctx, secrets, newMatchingLabelsSelector(options, true))
	if err != nil {
		return nil, err
	}
	for _, secret := range secrets.Items {
		if cluster, err := NewClusterFromSecret(secret.DeepCopy()); err == nil {
			clusters.Items = append(clusters.Items, *cluster)
		}
	}

	managedClusters := &ocmclusterv1.ManagedClusterList{}
	err = singleton.GetKubeClient().List(ctx, managedClusters, newMatchingLabelsSelector(options, false))
	if err != nil && !meta.IsNoMatchError(err) {
		return nil, err
	}
	for _, managedCluster := range managedClusters.Items {
		if cluster, err := NewClusterFromManagedCluster(managedCluster.DeepCopy()); err == nil {
			clusters.Items = append(clusters.Items, *cluster)
		}
	}
	return clusters, nil
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
	}
	cluster.Spec.Accepted = true
	cluster.Spec.Endpoint = ClusterBlankEndpoint
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
	if !metav1.HasLabel(secret.ObjectMeta, clustergatewaycommon.LabelKeyClusterCredentialType) {
		return nil, NewInvalidClusterSecretError()
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

// matchingLabelsSelector filters the list/delete operation of cluster list
type matchingLabelsSelector struct {
	selector              labels.Selector
	requireCredentialType bool
}

// ApplyToList applies this configuration to the given list options.
func (m matchingLabelsSelector) ApplyToList(opts *client.ListOptions) {
	opts.LabelSelector = m.selector
	if opts.LabelSelector == nil {
		opts.LabelSelector = labels.NewSelector()
	}
	if m.requireCredentialType {
		r, _ := labels.NewRequirement(clustergatewaycommon.LabelKeyClusterCredentialType, selection.Exists, nil)
		opts.LabelSelector = opts.LabelSelector.Add(*r)
	}
	opts.Namespace = StorageNamespace
}

func newMatchingLabelsSelector(options *metainternalversion.ListOptions, requireCredentialType bool) matchingLabelsSelector {
	sel := matchingLabelsSelector{requireCredentialType: requireCredentialType}
	if options != nil {
		sel.selector = options.LabelSelector
	}
	return sel
}
