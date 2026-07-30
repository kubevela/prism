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

package dynamicresource

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/dynamic"

	"github.com/kubevela/pkg/util/singleton"
)

type DynamicResource struct {
	Uns *unstructured.Unstructured

	codec Codec
}

// uns returns in.Uns, lazily initializing it if nil. Field-manager internals
// construct a DynamicResource's zero value via reflection (bypassing New()),
// so every accessor must tolerate a nil Uns rather than panic.
func (in *DynamicResource) uns() *unstructured.Unstructured {
	if in.Uns == nil {
		in.Uns = &unstructured.Unstructured{}
	}
	return in.Uns
}

func (in *DynamicResource) GetNamespace() string {
	return in.uns().GetNamespace()
}

func (in *DynamicResource) SetNamespace(namespace string) {
	in.uns().SetNamespace(namespace)
}

func (in *DynamicResource) GetName() string {
	return in.uns().GetName()
}

func (in *DynamicResource) SetName(name string) {
	in.uns().SetName(name)
}

func (in *DynamicResource) GetGenerateName() string {
	return in.uns().GetGenerateName()
}

func (in *DynamicResource) SetGenerateName(name string) {
	in.uns().SetGenerateName(name)
}

func (in *DynamicResource) GetUID() types.UID {
	return in.uns().GetUID()
}

func (in *DynamicResource) SetUID(uid types.UID) {
	in.uns().SetUID(uid)
}

func (in *DynamicResource) GetResourceVersion() string {
	return in.uns().GetResourceVersion()
}

func (in *DynamicResource) SetResourceVersion(version string) {
	in.uns().SetResourceVersion(version)
}

func (in *DynamicResource) GetGeneration() int64 {
	return in.uns().GetGeneration()
}

func (in *DynamicResource) SetGeneration(generation int64) {
	in.uns().SetGeneration(generation)
}

func (in *DynamicResource) GetSelfLink() string {
	return in.uns().GetSelfLink()
}

func (in *DynamicResource) SetSelfLink(selfLink string) {
	in.uns().SetSelfLink(selfLink)
}

func (in *DynamicResource) GetCreationTimestamp() metav1.Time {
	return in.uns().GetCreationTimestamp()
}

func (in *DynamicResource) SetCreationTimestamp(timestamp metav1.Time) {
	in.uns().SetCreationTimestamp(timestamp)
}

func (in *DynamicResource) GetDeletionTimestamp() *metav1.Time {
	return in.uns().GetDeletionTimestamp()
}

func (in *DynamicResource) SetDeletionTimestamp(timestamp *metav1.Time) {
	in.uns().SetDeletionTimestamp(timestamp)
}

func (in *DynamicResource) GetDeletionGracePeriodSeconds() *int64 {
	return in.uns().GetDeletionGracePeriodSeconds()
}

func (in *DynamicResource) SetDeletionGracePeriodSeconds(i *int64) {
	in.uns().SetDeletionGracePeriodSeconds(i)
}

func (in *DynamicResource) GetLabels() map[string]string {
	return in.uns().GetLabels()
}

func (in *DynamicResource) SetLabels(labels map[string]string) {
	in.uns().SetLabels(labels)
}

func (in *DynamicResource) GetAnnotations() map[string]string {
	return in.uns().GetAnnotations()
}

func (in *DynamicResource) SetAnnotations(annotations map[string]string) {
	in.uns().SetAnnotations(annotations)
}

func (in *DynamicResource) GetFinalizers() []string {
	return in.uns().GetFinalizers()
}

func (in *DynamicResource) SetFinalizers(finalizers []string) {
	in.uns().SetFinalizers(finalizers)
}

func (in *DynamicResource) GetOwnerReferences() []metav1.OwnerReference {
	return in.uns().GetOwnerReferences()
}

func (in *DynamicResource) SetOwnerReferences(references []metav1.OwnerReference) {
	in.uns().SetOwnerReferences(references)
}

func (in *DynamicResource) GetManagedFields() []metav1.ManagedFieldsEntry {
	return in.uns().GetManagedFields()
}

func (in *DynamicResource) SetManagedFields(managedFields []metav1.ManagedFieldsEntry) {
	in.uns().SetManagedFields(managedFields)
}

var _ runtime.Object = &DynamicResource{}

var _ rest.Storage = &DynamicResource{}
var _ rest.Scoper = &DynamicResource{}
var _ rest.Getter = &DynamicResource{}
var _ rest.CreaterUpdater = &DynamicResource{}
var _ rest.Patcher = &DynamicResource{}
var _ rest.GracefulDeleter = &DynamicResource{}
var _ rest.Lister = &DynamicResource{}
var _ rest.GroupVersionKindProvider = &DynamicResource{}

var _ metav1.Object = &DynamicResource{}

type dynamicResourceObjectKind struct {
	Typer
}

func (in *dynamicResourceObjectKind) SetGroupVersionKind(kind schema.GroupVersionKind) {}

func (in *dynamicResourceObjectKind) GroupVersionKind() schema.GroupVersionKind {
	if in.Typer != nil {
		return in.Typer.GroupVersionKind()
	}
	return schema.GroupVersionKind{}
}

var _ schema.ObjectKind = &dynamicResourceObjectKind{}

func (in *DynamicResource) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return containingGV.WithKind(in.codec.Source().Kind())
}

func (in *DynamicResource) GroupVersionResource() schema.GroupVersionResource {
	return in.codec.Source().GroupVersionResource()
}

func (in *DynamicResource) GetObjectKind() schema.ObjectKind {
	objKind := &dynamicResourceObjectKind{nil}
	if in.codec != nil {
		objKind.Typer = in.codec.Source()
	}
	return objKind
}

func (in *DynamicResource) GetObjectMeta() *metav1.ObjectMeta {
	uns := in.uns()
	return &metav1.ObjectMeta{
		Name:                       uns.GetName(),
		GenerateName:               uns.GetGenerateName(),
		Namespace:                  uns.GetNamespace(),
		UID:                        uns.GetUID(),
		ResourceVersion:            uns.GetResourceVersion(),
		Generation:                 uns.GetGeneration(),
		CreationTimestamp:          uns.GetCreationTimestamp(),
		DeletionTimestamp:          uns.GetDeletionTimestamp(),
		DeletionGracePeriodSeconds: uns.GetDeletionGracePeriodSeconds(),
		Labels:                     uns.GetLabels(),
		Annotations:                uns.GetAnnotations(),
		OwnerReferences:            uns.GetOwnerReferences(),
		Finalizers:                 uns.GetFinalizers(),
		ManagedFields:              uns.GetManagedFields(),
	}
}

func (in *DynamicResource) NamespaceScoped() bool {
	return in.codec.Source().Namespaced()
}

func (in *DynamicResource) New() runtime.Object {
	dr := &DynamicResource{
		Uns:   &unstructured.Unstructured{},
		codec: in.codec,
	}
	if dr.codec != nil {
		dr.Uns.SetGroupVersionKind(dr.codec.Source().GroupVersionKind())
	}
	return dr
}

func (in *DynamicResource) NewList() runtime.Object {
	drs := &DynamicResourceList{
		UnsList: &unstructured.UnstructuredList{},
		codec:   in.codec,
	}
	if drs.codec != nil {
		drs.UnsList.SetGroupVersionKind(drs.codec.Source().GroupVersionKindList())
	}
	return drs
}

func (in *DynamicResource) Destroy() {}

func (in *DynamicResource) GetGroupVersionResource() schema.GroupVersionResource {
	return in.codec.Source().GroupVersionResource()
}

func (in *DynamicResource) GetGroupVersionKind() schema.GroupVersionKind {
	return in.codec.Source().GroupVersionKind()
}

func (in *DynamicResource) GetGroupVersion() schema.GroupVersion {
	return in.codec.Source().GroupVersion()
}

// GetSingularName implements SingularNameProvider
func (in *DynamicResource) GetSingularName() string {
	return strings.ToLower(in.codec.Source().Kind())
}

func (in *DynamicResource) IsStorageVersion() bool {
	return true
}

func (in *DynamicResource) DeepCopyObject() runtime.Object {
	return &DynamicResource{
		Uns:   in.uns().DeepCopy(),
		codec: in.codec,
	}
}

func (in *DynamicResource) resourceInterface(ctx context.Context) dynamic.ResourceInterface {
	var ri dynamic.ResourceInterface = singleton.DynamicClient.Get().Resource(in.codec.Target().GroupVersionResource())
	if in.codec.Target().Namespaced() {
		ri = ri.(dynamic.NamespaceableResourceInterface).Namespace(request.NamespaceValue(ctx))
	}
	return ri
}

func (in *DynamicResource) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	obj, err := in.resourceInterface(ctx).Get(ctx, name, *options)
	if err != nil {
		return nil, err
	}
	decoded, err := in.codec.Decode(obj)
	return decoded, err
}

func (in *DynamicResource) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	o := obj.(*DynamicResource)
	encoded, err := in.codec.Encode(o.uns())
	if err != nil {
		return nil, err
	}
	created, err := in.resourceInterface(ctx).Create(ctx, encoded, *options)
	if err != nil {
		return nil, err
	}
	decoded, err := in.codec.Decode(created)
	return decoded, err
}

func (in *DynamicResource) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	decoded, err := in.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}
	updated, err := objInfo.UpdatedObject(ctx, decoded)
	if err != nil {
		return nil, false, err
	}
	encoded, err := in.codec.Encode(updated.(*DynamicResource).uns())
	if err != nil {
		return nil, false, err
	}
	encoded, err = in.resourceInterface(ctx).Update(ctx, encoded, *options)
	if err != nil {
		return nil, false, err
	}
	decoded, err = in.codec.Decode(encoded)
	return decoded, false, err
}

func (in *DynamicResource) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	obj, err := in.resourceInterface(ctx).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}
	if err = in.resourceInterface(ctx).Delete(ctx, name, *options); err != nil {
		return nil, false, err
	}
	decoded, err := in.codec.Decode(obj)
	return decoded, false, err
}

func (in *DynamicResource) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	v1Options := &metav1.ListOptions{}
	if err := internalversion.Convert_internalversion_ListOptions_To_v1_ListOptions(options, v1Options, nil); err != nil {
		return nil, err
	}
	objs, err := in.resourceInterface(ctx).List(ctx, *v1Options)
	if err != nil {
		return nil, err
	}
	decoded := &unstructured.UnstructuredList{}
	decoded.SetGroupVersionKind(in.codec.Source().GroupVersionKindList())
	for i := range objs.Items {
		d, err := in.codec.Decode(&objs.Items[i])
		if err != nil {
			return nil, err
		}
		decoded.Items = append(decoded.Items, *d)
	}
	return decoded, nil
}

func (in *DynamicResource) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	// TODO
	t := &metav1.Table{
		ColumnDefinitions: []metav1.TableColumnDefinition{{
			Name: "Name",
			Type: "string",
		}, {
			Name: "CreateAt",
			Type: "dateTime",
		}},
		Rows: []metav1.TableRow{},
	}
	var dat []*unstructured.Unstructured
	switch obj := object.(type) {
	case *unstructured.Unstructured:
		dat = append(dat, obj)
	case *unstructured.UnstructuredList:
		for i := range obj.Items {
			dat = append(dat, &obj.Items[i])
		}
	default:
		return nil, fmt.Errorf("unknown type %T", object)
	}
	for _, u := range dat {
		row := metav1.TableRow{
			Object: runtime.RawExtension{Object: u},
		}
		row.Cells = []interface{}{u.GetName(), u.GetCreationTimestamp()}
		t.Rows = append(t.Rows, row)
	}
	return t, nil
}

func (in *DynamicResource) MarshalJSON() ([]byte, error) {
	return json.Marshal(in.uns())
}

func (in *DynamicResource) UnmarshalJSON(bs []byte) error {
	in.Uns = &unstructured.Unstructured{}
	return in.Uns.UnmarshalJSON(bs)
}
