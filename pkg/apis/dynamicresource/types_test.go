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

package dynamicresource_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/pointer"

	"github.com/kubevela/prism/pkg/apis/dynamicresource"
)

const (
	encoderTemplate = `
		parameter: {
	        apiVersion: "test.oam.dev/v1alpha2"
	        kind: "Tester"
	        metadata: {...}
	    }
	    output: {
	        apiVersion: "v1"
	        kind: "ConfigMap"
	        metadata: parameter.metadata
			data: {}
	    }
	`
	decoderTemplate = `
		parameter: {
	        apiVersion: "v1"
	        kind: "ConfigMap"
	        metadata: {...}
			data: {...}
	    }
	    output: {
	        apiVersion: "test.oam.dev/v1alpha2"
	        kind: "Tester"
	        metadata: parameter.metadata
	    }
	`
)

func newDynamicResource() (*dynamicresource.DynamicResource, error) {
	return dynamicresource.NewDynamicResourceWithCodec(encoderTemplate, decoderTemplate)
}

func testSetterAndGetter[T any](setter func(T), getter func() T, val T) {
	setter(val)
	Ω(getter()).Should(Equal(val))
}

var _ = Describe("Test dynamic resource", func() {
	It("Test types", func() {
		store, err := newDynamicResource()
		Ω(err).To(Succeed())
		dr := store.New().(*dynamicresource.DynamicResource)
		testSetterAndGetter(dr.SetNamespace, dr.GetNamespace, "ns")
		testSetterAndGetter(dr.SetName, dr.GetName, "name")
		testSetterAndGetter(dr.SetUID, dr.GetUID, "name")
		testSetterAndGetter(dr.SetGenerateName, dr.GetGenerateName, "gn")
		testSetterAndGetter(dr.SetResourceVersion, dr.GetResourceVersion, "rv")
		testSetterAndGetter(dr.SetGeneration, dr.GetGeneration, 1)
		testSetterAndGetter(dr.SetSelfLink, dr.GetSelfLink, "sl")
		t := time.Date(2022, 1, 1, 0, 0, 0, 0, time.Local)
		testSetterAndGetter(dr.SetCreationTimestamp, dr.GetCreationTimestamp, metav1.Time{Time: t})
		testSetterAndGetter(dr.SetDeletionTimestamp, dr.GetDeletionTimestamp, &metav1.Time{Time: t})
		testSetterAndGetter(dr.SetDeletionGracePeriodSeconds, dr.GetDeletionGracePeriodSeconds, pointer.Int64(1))
		testSetterAndGetter(dr.SetLabels, dr.GetLabels, map[string]string{"x": "y"})
		testSetterAndGetter(dr.SetAnnotations, dr.GetAnnotations, map[string]string{"x": "x"})
		testSetterAndGetter(dr.SetFinalizers, dr.GetFinalizers, []string{"f"})
		testSetterAndGetter(dr.SetOwnerReferences, dr.GetOwnerReferences, []metav1.OwnerReference{{Kind: "k"}})
		testSetterAndGetter(dr.SetManagedFields, dr.GetManagedFields, []metav1.ManagedFieldsEntry{{Manager: "m"}})

		Ω(dr.GroupVersionKind(schema.GroupVersion{})).To(Equal(schema.GroupVersionKind{Kind: "Tester"}))
		gv := schema.GroupVersion{Group: "test.oam.dev", Version: "v1alpha2"}
		gvr := gv.WithResource("testers")
		gvk := gv.WithKind("Tester")
		Ω(dr.GroupVersionResource()).To(Equal(gvr))
		objKind := dr.GetObjectKind()
		objKind.SetGroupVersionKind(schema.GroupVersionKind{})
		Ω(objKind.GroupVersionKind()).To(Equal(gvk))
		Ω(dr.NamespaceScoped()).To(BeTrue())
		defer dr.Destroy()
		Ω(dr.GetGroupVersionResource()).To(Equal(gvr))
		Ω(dr.GetGroupVersionKind()).To(Equal(gvk))
		Ω(dr.GetGroupVersion()).To(Equal(gv))
		Ω(dr.IsStorageVersion()).To(BeTrue())

		Ω(dr.DeepCopyObject().(*dynamicresource.DynamicResource).GetObjectMeta()).To(Equal(dr.GetObjectMeta()))
		bs, err := dr.MarshalJSON()
		Ω(err).To(Succeed())
		ndr := dr.New().(*dynamicresource.DynamicResource)
		Ω(ndr.UnmarshalJSON(bs)).To(Succeed())
		Ω(ndr.GetObjectMeta()).To(Equal(dr.GetObjectMeta()))

		drs := dr.NewList().(*dynamicresource.DynamicResourceList)
		drs.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{})
		Ω(drs.GetObjectKind().GroupVersionKind()).To(Equal(gv.WithKind("TesterList")))
		drs.UnsList.Items = make([]unstructured.Unstructured, 1)
		Ω(len(drs.DeepCopyObject().(*dynamicresource.DynamicResourceList).UnsList.Items)).To(Equal(1))
		ndrs := dr.NewList().(*dynamicresource.DynamicResourceList)
		bs, err = drs.MarshalJSON()
		Ω(err).To(Succeed())
		Ω(ndrs.UnmarshalJSON(bs)).To(Succeed())
		Ω(len(ndrs.DeepCopyObject().(*dynamicresource.DynamicResourceList).UnsList.Items)).To(Equal(1))
	})

	It("Test CURD API", func() {
		store, err := newDynamicResource()
		Ω(err).To(Succeed())

		ctx := request.WithNamespace(context.Background(), "default")
		By("Create")
		dr := store.New().(*dynamicresource.DynamicResource)
		dr.SetName("example")
		dr.SetLabels(map[string]string{"key": "val"})
		_, err = store.Create(ctx, dr, nil, &metav1.CreateOptions{})
		Ω(err).To(Succeed())

		By("Get")
		obj, err := store.Get(ctx, "example", &metav1.GetOptions{})
		Ω(err).To(Succeed())
		uns := obj.(*unstructured.Unstructured)
		Ω(uns.GetLabels()["key"]).To(Equal("val"))
		_, err = store.ConvertToTable(ctx, uns, nil)
		Ω(err).To(Succeed())

		By("Update")
		dr.SetLabels(map[string]string{"key": "value"})
		_, _, err = store.Update(ctx, dr.GetName(), rest.DefaultUpdatedObjectInfo(dr), nil, nil, false, &metav1.UpdateOptions{})
		Ω(err).To(Succeed())

		By("List")
		objs, err := store.List(ctx, &internalversion.ListOptions{})
		Ω(err).To(Succeed())
		unsList := objs.(*unstructured.UnstructuredList)
		Ω(len(unsList.Items)).To(Equal(1))
		Ω(unsList.Items[0].GetLabels()["key"]).To(Equal("value"))
		_, err = store.ConvertToTable(ctx, unsList, nil)
		Ω(err).To(Succeed())

		By("Delete")
		_, _, err = store.Delete(ctx, "example", nil, &metav1.DeleteOptions{})
		Ω(err).To(Succeed())
		_, err = store.Get(ctx, "example", &metav1.GetOptions{})
		Ω(kerrors.IsNotFound(err)).To(BeTrue())
	})
})
