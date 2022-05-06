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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/endpoints/request"
)

var _ = Describe("Test ApplicationResourceTracker API", func() {

	It("Test usa of ResourceHandlerProvider for ApplicationResourceTracker", func() {
		By("Test create handler")
		provider := NewResourceHandlerProvider(cfg)
		_s, err := provider(nil, nil)
		Ω(err).To(Succeed())
		s, ok := _s.(*storage)
		Ω(ok).To(BeTrue())

		By("Test meta info")
		Ω(s.New()).To(Equal(&ApplicationResourceTracker{}))
		Ω(s.NamespaceScoped()).To(BeTrue())
		Ω(s.ShortNames()).To(ContainElement("apprt"))
		Ω(s.NewList()).To(Equal(&ApplicationResourceTrackerList{}))

		ctx := context.Background()

		By("Create RT")
		createRt := func(name, ns, val string) *unstructured.Unstructured {
			rt := &unstructured.Unstructured{}
			rt.SetGroupVersionKind(ResourceTrackerGroupVersionKind)
			rt.SetName(name + "-" + ns)
			rt.SetLabels(map[string]string{
				labelAppNamespace: ns,
				"key":             val,
			})
			Ω(k8sClient.Create(ctx, rt)).To(Succeed())
			return rt
		}
		createRt("app-1", "example", "x")
		createRt("app-2", "example", "y")
		createRt("app-1", "default", "x")
		createRt("app-2", "default", "x")
		createRt("app-3", "default", "x")

		By("Test Get")
		_appRt1, err := s.Get(request.WithNamespace(ctx, "default"), "app-1", nil)
		Ω(err).To(Succeed())
		appRt1, ok := _appRt1.(*ApplicationResourceTracker)
		Ω(ok).To(BeTrue())
		Ω(appRt1.GetLabels()["key"]).To(Equal("x"))
		_, err = s.Get(request.WithNamespace(ctx, "no"), "app-1", nil)
		Ω(errors.IsNotFound(err)).To(BeTrue())

		By("Test List")
		_appRts1, err := s.List(request.WithNamespace(ctx, "example"), nil)
		Ω(err).To(Succeed())
		appRts1, ok := _appRts1.(*ApplicationResourceTrackerList)
		Ω(ok).To(BeTrue())
		Ω(len(appRts1.Items)).To(Equal(2))

		_appRts2, err := s.List(ctx, &metainternalversion.ListOptions{LabelSelector: labels.SelectorFromValidatedSet(map[string]string{"key": "x"})})
		Ω(err).To(Succeed())
		appRts2, ok := _appRts2.(*ApplicationResourceTrackerList)
		Ω(ok).To(BeTrue())
		Ω(len(appRts2.Items)).To(Equal(4))

		_appRts3, err := s.List(request.WithNamespace(ctx, "default"), &metainternalversion.ListOptions{LabelSelector: labels.SelectorFromValidatedSet(map[string]string{"key": "x"})})
		Ω(err).To(Succeed())
		appRts3, ok := _appRts3.(*ApplicationResourceTrackerList)
		Ω(ok).To(BeTrue())
		Ω(len(appRts3.Items)).To(Equal(3))
	})

})
