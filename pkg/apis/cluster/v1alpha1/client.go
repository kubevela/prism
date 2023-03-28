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
	"sort"

	clustergatewaycommon "github.com/oam-dev/cluster-gateway/pkg/common"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	apitypes "k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/strings/slices"
	ocmclusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubevela/pkg/util/apiserver"
)

// ClusterClient client for reading cluster information
// +kubebuilder:object:generate=false
type ClusterClient interface {
	Get(ctx context.Context, name string) (*Cluster, error)
	List(ctx context.Context, options ...client.ListOption) (*ClusterList, error)
}

type clusterClient struct {
	client.Client
}

// NewClusterClient create a client for accessing cluster
func NewClusterClient(cli client.Client) ClusterClient {
	return &clusterClient{Client: cli}
}

func (c *clusterClient) Get(ctx context.Context, name string) (*Cluster, error) {
	if name == ClusterLocalName {
		return NewLocalCluster(), nil
	}
	key := apitypes.NamespacedName{Name: name, Namespace: StorageNamespace}
	var cluster *Cluster
	secret := &corev1.Secret{}
	err := c.Client.Get(ctx, key, secret)
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
	err = c.Client.Get(ctx, key, managedCluster)
	var managedClusterErr error
	if err == nil {
		if cluster, managedClusterErr = NewClusterFromManagedCluster(managedCluster); managedClusterErr == nil {
			return cluster, nil
		}
	}

	if err != nil && !apierrors.IsNotFound(err) && !meta.IsNoMatchError(err) && !runtime.IsNotRegisteredError(err) {
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

func (c *clusterClient) List(ctx context.Context, options ...client.ListOption) (*ClusterList, error) {
	opts := apiserver.NewListOptions(options...)
	local := NewLocalCluster()
	clusters := &ClusterList{Items: []Cluster{*local}}

	secrets := &corev1.SecretList{}
	err := c.Client.List(ctx, secrets, clusterSelector{Selector: opts.LabelSelector, RequireCredentialType: true})
	if err != nil {
		return nil, err
	}
	for _, secret := range secrets.Items {
		if cluster, err := NewClusterFromSecret(secret.DeepCopy()); err == nil {
			clusters.Items = append(clusters.Items, *cluster)
		}
	}

	managedClusters := &ocmclusterv1.ManagedClusterList{}
	err = c.Client.List(ctx, managedClusters, clusterSelector{Selector: opts.LabelSelector, RequireCredentialType: false, IgnoreNamespace: true})
	if err != nil && !meta.IsNoMatchError(err) && !runtime.IsNotRegisteredError(err) {
		return nil, err
	}
	for _, managedCluster := range managedClusters.Items {
		if !clusters.HasCluster(managedCluster.Name) {
			if cluster, err := NewClusterFromManagedCluster(managedCluster.DeepCopy()); err == nil {
				clusters.Items = append(clusters.Items, *cluster)
			}
		}
	}

	// filter clusters
	var items []Cluster
	for _, cluster := range clusters.Items {
		if opts.LabelSelector == nil || opts.LabelSelector.Matches(labels.Set(cluster.GetLabels())) {
			items = append(items, cluster)
		}
	}
	clusters.Items = items

	// sort clusters
	sort.Slice(clusters.Items, func(i, j int) bool {
		if clusters.Items[i].Name == ClusterLocalName {
			return true
		} else if clusters.Items[j].Name == ClusterLocalName {
			return false
		} else {
			return clusters.Items[i].CreationTimestamp.After(clusters.Items[j].CreationTimestamp.Time)
		}
	})
	return clusters, nil
}

// clusterSelector filters the list/delete operation of cluster list
type clusterSelector struct {
	Selector              labels.Selector
	RequireCredentialType bool
	IgnoreNamespace       bool
}

// ApplyToList applies this configuration to the given list options.
func (m clusterSelector) ApplyToList(opts *client.ListOptions) {
	opts.LabelSelector = labels.NewSelector()
	if m.Selector != nil {
		requirements, _ := m.Selector.Requirements()
		for _, r := range requirements {
			if !slices.Contains([]string{LabelClusterControlPlane}, r.Key()) {
				opts.LabelSelector = opts.LabelSelector.Add(r)
			}
		}
	}
	if m.RequireCredentialType {
		r, _ := labels.NewRequirement(clustergatewaycommon.LabelKeyClusterCredentialType, selection.Exists, nil)
		opts.LabelSelector = opts.LabelSelector.Add(*r)
	}
	if !m.IgnoreNamespace {
		opts.Namespace = StorageNamespace
	}
}
