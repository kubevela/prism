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
	"fmt"

	clusterv1alpha1 "github.com/oam-dev/cluster-gateway/pkg/apis/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

// Cluster is an extension model for cluster underlying secrets/ManagedClusters
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterSpec `json:"spec,omitempty"`
}

// ClusterSpec spec of cluster
type ClusterSpec struct {
	Alias          string                         `json:"alias,omitempty"`
	Accepted       bool                           `json:"accepted,omitempty"`
	Endpoint       string                         `json:"endpoint,omitempty"`
	CredentialType clusterv1alpha1.CredentialType `json:"credential-type,omitempty"`
}

// ClusterList list for Cluster
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Cluster `json:"items"`
}

var _ resource.Object = &Cluster{}

// GetObjectMeta returns the object meta reference.
func (in *Cluster) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// NamespaceScoped returns if the object must be in a namespace.
func (in *Cluster) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (in *Cluster) New() runtime.Object {
	return &Cluster{}
}

// NewList return a new list instance of the resource
func (in *Cluster) NewList() runtime.Object {
	return &ClusterList{}
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
func (in *Cluster) GetGroupVersionResource() schema.GroupVersionResource {
	return GroupVersion.WithResource(ClusterResource)
}

// IsStorageVersion returns true if the object is also the internal version
func (in *Cluster) IsStorageVersion() bool {
	return true
}

// ShortNames delivers a list of short names for a resource.
func (in *Cluster) ShortNames() []string {
	return []string{"vc", "vela-cluster", "vela-clusters"}
}

// GetFullName returns the name with alias
func (in *Cluster) GetFullName() string {
	if in.Spec.Alias == "" {
		return in.GetName()
	}
	return fmt.Sprintf("%s (%s)", in.GetName(), in.Spec.Alias)
}
