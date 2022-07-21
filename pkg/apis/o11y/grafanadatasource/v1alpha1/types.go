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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"

	"github.com/kubevela/prism/pkg/util/singleton"
)

// GrafanaDatasource is a reflection api for Grafana Datasource
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GrafanaDatasource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Spec runtime.RawExtension `json:"spec,omitempty"`
}

// GrafanaDatasourceList list for GrafanaDatasource
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GrafanaDatasourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []GrafanaDatasource `json:"items"`
}

var _ resource.Object = &GrafanaDatasource{}
var _ rest.Getter = &GrafanaDatasource{}
var _ rest.CreaterUpdater = &GrafanaDatasource{}
var _ rest.Patcher = &GrafanaDatasource{}
var _ rest.GracefulDeleter = &GrafanaDatasource{}

// GetObjectMeta returns the object meta reference.
func (in *GrafanaDatasource) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// NamespaceScoped returns if the object must be in a namespace.
func (in *GrafanaDatasource) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (in *GrafanaDatasource) New() runtime.Object {
	return &GrafanaDatasource{}
}

// NewList return a new list instance of the resource
func (in *GrafanaDatasource) NewList() runtime.Object {
	return &GrafanaDatasourceList{}
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
func (in *GrafanaDatasource) GetGroupVersionResource() schema.GroupVersionResource {
	return GroupVersion.WithResource(GrafanaDatasourceResource)
}

// IsStorageVersion returns true if the object is also the internal version
func (in *GrafanaDatasource) IsStorageVersion() bool {
	return true
}

// ShortNames delivers a list of short names for a resource.
func (in *GrafanaDatasource) ShortNames() []string {
	return []string{"gds", "datasource", "grafana-datasource"}
}

// Get finds a resource in the storage by name and returns it.
func (in *GrafanaDatasource) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return NewGrafanaDatasourceClient(singleton.GetKubeClient()).Get(ctx, name)
}

func (in *GrafanaDatasource) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	datasource := obj.(*GrafanaDatasource)
	return datasource, NewGrafanaDatasourceClient(singleton.GetKubeClient()).Create(ctx, datasource)
}

func (in *GrafanaDatasource) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	cli := NewGrafanaDatasourceClient(singleton.GetKubeClient())
	datasource, err := cli.Get(ctx, name)
	if err != nil {
		return nil, false, err
	}
	obj, err := objInfo.UpdatedObject(ctx, datasource)
	if err != nil {
		return nil, false, err
	}
	datasource = obj.(*GrafanaDatasource)
	err = cli.Update(ctx, datasource)
	return datasource, false, err
}

func (in *GrafanaDatasource) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	cli := NewGrafanaDatasourceClient(singleton.GetKubeClient())
	datasource, err := cli.Get(ctx, name)
	if err != nil {
		return nil, false, err
	}
	return datasource, true, cli.Delete(ctx, datasource)
}

// TODO support list datasource
