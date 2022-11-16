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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/endpoints/request"
	apirest "k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubevela/prism/pkg/util/singleton"
)

// ApplicationResourceTracker is an extension model for ResourceTracker
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ApplicationResourceTracker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Spec runtime.RawExtension `json:"spec,omitempty"`
}

// ApplicationResourceTrackerList list for ApplicationResourceTracker
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ApplicationResourceTrackerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ApplicationResourceTracker `json:"items"`
}

var _ resource.Object = &ApplicationResourceTracker{}

// GetObjectMeta returns the object meta reference.
func (in *ApplicationResourceTracker) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// NamespaceScoped returns if the object must be in a namespace.
func (in *ApplicationResourceTracker) NamespaceScoped() bool {
	return true
}

// New returns a new instance of the resource
func (in *ApplicationResourceTracker) New() runtime.Object {
	return &ApplicationResourceTracker{}
}

// Destroy .
func (in *ApplicationResourceTracker) Destroy() {}

// NewList return a new list instance of the resource
func (in *ApplicationResourceTracker) NewList() runtime.Object {
	return &ApplicationResourceTrackerList{}
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
func (in *ApplicationResourceTracker) GetGroupVersionResource() schema.GroupVersionResource {
	return GroupVersion.WithResource(ApplicationResourceTrackerResource)
}

// IsStorageVersion returns true if the object is also the internal version
func (in *ApplicationResourceTracker) IsStorageVersion() bool {
	return true
}

// ShortNames delivers a list of short names for a resource.
func (in *ApplicationResourceTracker) ShortNames() []string {
	return []string{"apprt"}
}

// Get finds a resource in the storage by name and returns it.
func (in *ApplicationResourceTracker) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	rt := &unstructured.Unstructured{}
	rt.SetGroupVersionKind(ResourceTrackerGroupVersionKind)
	ns := request.NamespaceValue(ctx)
	if err := singleton.KubeClient.Get().Get(ctx, types.NamespacedName{Name: name + "-" + ns}, rt); err != nil {
		return nil, err
	}
	return NewApplicationResourceTrackerFromResourceTracker(rt)
}

// List selects resources in the storage which match to the selector. 'options' can be nil.
func (in *ApplicationResourceTracker) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	rts := &unstructured.UnstructuredList{}
	rts.SetGroupVersionKind(ResourceTrackerGroupVersionKind)
	ns := request.NamespaceValue(ctx)
	sel := matchingLabelsSelector{namespace: ns}
	if options != nil {
		sel.selector = options.LabelSelector
	}
	if err := singleton.KubeClient.Get().List(ctx, rts, sel); err != nil {
		return nil, err
	}
	appRts := &ApplicationResourceTrackerList{}
	for _, rt := range rts.Items {
		appRt, err := NewApplicationResourceTrackerFromResourceTracker(rt.DeepCopy())
		if err != nil {
			return nil, err
		}
		appRts.Items = append(appRts.Items, *appRt)
	}
	return appRts, nil
}

// ConvertToTable convert resource to table
func (in *ApplicationResourceTracker) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return apirest.NewDefaultTableConvertor(ApplicationResourceTrackerGroupResource).ConvertToTable(ctx, object, tableOptions)
}

type matchingLabelsSelector struct {
	selector  labels.Selector
	namespace string
}

// ApplyToList applies this configuration to the given list options.
func (m matchingLabelsSelector) ApplyToList(opts *client.ListOptions) {
	opts.LabelSelector = m.selector
	if opts.LabelSelector == nil {
		opts.LabelSelector = labels.NewSelector()
	}
	if m.namespace != "" {
		sel := labels.SelectorFromValidatedSet(map[string]string{labelAppNamespace: m.namespace})
		r, _ := sel.Requirements()
		opts.LabelSelector = opts.LabelSelector.Add(r...)
	}
}
