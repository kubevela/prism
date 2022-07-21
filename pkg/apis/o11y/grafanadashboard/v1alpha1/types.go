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

// GrafanaDashboard is a reflection api for Grafana Datasource
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GrafanaDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Spec runtime.RawExtension `json:"spec,omitempty"`
}

// GrafanaDashboardList list for GrafanaDashboard
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GrafanaDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []GrafanaDashboard `json:"items"`
}

var _ resource.Object = &GrafanaDashboard{}
var _ rest.Getter = &GrafanaDashboard{}
var _ rest.CreaterUpdater = &GrafanaDashboard{}
var _ rest.Patcher = &GrafanaDashboard{}
var _ rest.GracefulDeleter = &GrafanaDashboard{}

// GetObjectMeta returns the object meta reference.
func (in *GrafanaDashboard) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// NamespaceScoped returns if the object must be in a namespace.
func (in *GrafanaDashboard) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (in *GrafanaDashboard) New() runtime.Object {
	return &GrafanaDashboard{}
}

// NewList return a new list instance of the resource
func (in *GrafanaDashboard) NewList() runtime.Object {
	return &GrafanaDashboardList{}
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
func (in *GrafanaDashboard) GetGroupVersionResource() schema.GroupVersionResource {
	return GroupVersion.WithResource(GrafanaDashboardResource)
}

// IsStorageVersion returns true if the object is also the internal version
func (in *GrafanaDashboard) IsStorageVersion() bool {
	return true
}

// ShortNames delivers a list of short names for a resource.
func (in *GrafanaDashboard) ShortNames() []string {
	return []string{"gdb", "datasource", "grafana-datasource"}
}

// Get finds a resource in the storage by name and returns it.
func (in *GrafanaDashboard) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return NewGrafanaDashboardClient(singleton.GetKubeClient()).Get(ctx, name)
}

func (in *GrafanaDashboard) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	gdb := obj.(*GrafanaDashboard)
	return gdb, NewGrafanaDashboardClient(singleton.GetKubeClient()).Create(ctx, gdb)
}

func (in *GrafanaDashboard) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	cli := NewGrafanaDashboardClient(singleton.GetKubeClient())
	gdb, err := cli.Get(ctx, name)
	if err != nil {
		return nil, false, err
	}
	obj, err := objInfo.UpdatedObject(ctx, gdb)
	if err != nil {
		return nil, false, err
	}
	gdb = obj.(*GrafanaDashboard)
	err = cli.Update(ctx, gdb)
	return gdb, false, err
}

func (in *GrafanaDashboard) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	cli := NewGrafanaDashboardClient(singleton.GetKubeClient())
	gdb, err := cli.Get(ctx, name)
	if err != nil {
		return nil, false, err
	}
	return gdb, true, cli.Delete(ctx, gdb)
}

// TODO support list datasource
