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

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubevela/prism/pkg/util/singleton"
)

// Grafana defines the instance of grafana
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Grafana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GrafanaSpec `json:"spec,omitempty"`
}

// GrafanaSpec defines the spec for grafana instance
type GrafanaSpec struct {
	Endpoint string           `json:"endpoint"`
	Access   AccessCredential `json:"access"`
}

// BasicAuth defines the basic auth credential
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AccessCredential defines the access credential for the grafana api
type AccessCredential struct {
	*BasicAuth `json:",inline,omitempty"`
	Token      *string `json:"token,omitempty"`
}

// GrafanaList list for Grafana
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GrafanaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Grafana `json:"items"`
}

var _ resource.Object = &Grafana{}
var _ rest.Getter = &Grafana{}
var _ rest.Lister = &Grafana{}
var _ rest.CreaterUpdater = &Grafana{}
var _ rest.Patcher = &Grafana{}
var _ rest.GracefulDeleter = &Grafana{}

// GetObjectMeta returns the object meta reference.
func (in *Grafana) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// NamespaceScoped returns if the object must be in a namespace.
func (in *Grafana) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (in *Grafana) New() runtime.Object {
	return &Grafana{}
}

// NewList return a new list instance of the resource
func (in *Grafana) NewList() runtime.Object {
	return &GrafanaList{}
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
func (in *Grafana) GetGroupVersionResource() schema.GroupVersionResource {
	return GroupVersion.WithResource(GrafanaResource)
}

// IsStorageVersion returns true if the object is also the internal version
func (in *Grafana) IsStorageVersion() bool {
	return true
}

// ShortNames delivers a list of short names for a resource.
func (in *Grafana) ShortNames() []string {
	return []string{"gf", "grafana-instance"}
}

const (
	grafanaSecretNamePrefix            = "grafana."
	grafanaSecretEndpointAnnotationKey = "o11y.oam.dev/grafana-endpoint"
	grafanaSecretUsernameKey           = "username"
	grafanaSecretPasswordKey           = "password"
	grafanaSecretTokenKey              = "token"
)

// Get finds a resource in the storage by name and returns it.
func (in *Grafana) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return NewGrafanaClient(singleton.GetKubeClient()).Get(ctx, name)
}

// List selects resources in the storage which match to the selector. 'options' can be nil.
func (in *Grafana) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	sel := labels.NewSelector()
	if options != nil && options.LabelSelector != nil && !options.LabelSelector.Empty() {
		sel = options.LabelSelector
	}
	return NewGrafanaClient(singleton.GetKubeClient()).List(ctx, client.MatchingLabelsSelector{Selector: sel})
}

// Create creates a new version of a resource.
func (in *Grafana) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	grafana := obj.(*Grafana)
	return grafana, NewGrafanaClient(singleton.GetKubeClient()).Create(ctx, grafana)
}

// Update finds a resource in the storage and updates it.
func (in *Grafana) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	gfClient := NewGrafanaClient(singleton.GetKubeClient())
	grafana, err := gfClient.Get(ctx, name)
	if err != nil {
		return nil, false, err
	}
	obj, err := objInfo.UpdatedObject(ctx, grafana)
	if err != nil {
		return nil, false, err
	}
	grafana = obj.(*Grafana)
	err = gfClient.Update(ctx, grafana)
	return grafana, false, err
}

// Delete finds a resource in the storage and deletes it.
func (in *Grafana) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	gfClient := NewGrafanaClient(singleton.GetKubeClient())
	grafana, err := gfClient.Get(ctx, name)
	if err != nil {
		return nil, false, err
	}
	err = gfClient.Delete(ctx, grafana)
	return grafana, true, err
}

// TODO add access check subresource
