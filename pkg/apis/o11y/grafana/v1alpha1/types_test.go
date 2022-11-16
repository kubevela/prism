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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/utils/pointer"

	"github.com/kubevela/prism/pkg/apis/o11y/config"
	"github.com/kubevela/prism/pkg/util/singleton"
	testutil "github.com/kubevela/prism/test/util"
)

var _ = Describe("Test Grafana API", func() {

	BeforeEach(func() {
		Ω(testutil.CreateNamespace(config.ObservabilityNamespace)).To(Succeed())
	})

	AfterEach(func() {
		Ω(testutil.DeleteNamespace(config.ObservabilityNamespace)).To(Succeed())
	})

	It("Test Grafana API", func() {
		s := &Grafana{}
		By("Test meta info")
		By("Test meta info")
		Ω(s.New()).To(Equal(&Grafana{}))
		Ω(s.NamespaceScoped()).To(BeFalse())
		Ω(s.ShortNames()).To(ContainElement("gf"))
		Ω(s.GetGroupVersionResource().GroupVersion()).To(Equal(GroupVersion))
		Ω(s.GetGroupVersionResource().Resource).To(Equal(GrafanaResource))
		Ω(s.IsStorageVersion()).To(BeTrue())
		Ω(s.NewList()).To(Equal(&GrafanaList{}))

		ctx := context.Background()

		By("Test Create Grafana")
		g1 := &Grafana{
			ObjectMeta: metav1.ObjectMeta{Name: "example1", Labels: map[string]string{"key": "value"}},
			Spec:       GrafanaSpec{Endpoint: "1", Access: AccessCredential{Token: pointer.String("-")}},
		}
		g2 := &Grafana{
			ObjectMeta: metav1.ObjectMeta{Name: "example2", Labels: map[string]string{"key": "value"}},
			Spec:       GrafanaSpec{Endpoint: "2", Access: AccessCredential{BasicAuth: &BasicAuth{Username: "-", Password: "-"}}},
		}
		g3 := &Grafana{
			ObjectMeta: metav1.ObjectMeta{Name: "example3", Annotations: map[string]string{"key": "value"}},
			Spec:       GrafanaSpec{Endpoint: "3", Access: AccessCredential{BasicAuth: &BasicAuth{Username: "-", Password: "-"}}},
		}
		for _, g := range []*Grafana{g1, g2, g3} {
			_, err := s.Create(ctx, g, nil, nil)
			Ω(err).To(Succeed())
		}

		By("Create secret for distinguish test")
		secret1 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "grafana.bad1"},
		}
		secret2 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "grafana.bad2", Annotations: map[string]string{grafanaSecretEndpointAnnotationKey: "-"}},
		}
		secret3 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "bad3"},
		}
		for _, secret := range []*corev1.Secret{secret1, secret2, secret3} {
			secret.SetNamespace(config.ObservabilityNamespace)
			Ω(singleton.KubeClient.Get().Create(context.Background(), secret)).To(Succeed())
		}

		By("Test Get Grafana")
		obj, err := s.Get(ctx, "example1", nil)
		Ω(err).To(Succeed())
		grafana, ok := obj.(*Grafana)
		Ω(ok).To(BeTrue())
		Ω(grafana.Spec).To(Equal(g1.Spec))
		Ω(grafana.GetCredentialType()).To(Equal(GrafanaCredentialTypeBearerToken))
		obj, err = s.Get(ctx, "example2", nil)
		Ω(err).To(Succeed())
		grafana, ok = obj.(*Grafana)
		Ω(ok).To(BeTrue())
		Ω(grafana.Spec).To(Equal(g2.Spec))
		Ω(grafana.GetCredentialType()).To(Equal(GrafanaCredentialTypeBasicAuth))
		_, err = s.Get(ctx, "example4", nil)
		Ω(err).To(Satisfy(errors.IsNotFound))
		_, err = s.Get(ctx, "bad1", nil)
		Ω(err).To(Equal(NewEmptyEndpointGrafanaSecretError()))
		_, err = s.Get(ctx, "bad2", nil)
		Ω(err).To(Equal(NewEmptyCredentialGrafanaSecretError()))

		By("Test List Grafana")
		objs, err := s.List(ctx, nil)
		Ω(err).To(Succeed())
		grafanas, ok := objs.(*GrafanaList)
		Ω(ok).To(BeTrue())
		Ω(len(grafanas.Items)).To(Equal(3))
		objs, err = s.List(ctx, &metainternalversion.ListOptions{LabelSelector: labels.SelectorFromSet(map[string]string{"key": "value"})})
		Ω(err).To(Succeed())
		grafanas, ok = objs.(*GrafanaList)
		Ω(ok).To(BeTrue())
		Ω(len(grafanas.Items)).To(Equal(2))

		By("Test print table")
		_, err = s.ConvertToTable(ctx, grafana, nil)
		Ω(err).To(Succeed())
		_, err = s.ConvertToTable(ctx, grafanas, nil)
		Ω(err).To(Succeed())

		By("Test Update Grafana")
		obj, _, err = s.Update(ctx, "example3", rest.DefaultUpdatedObjectInfo(nil, func(ctx context.Context, newObj runtime.Object, oldObj runtime.Object) (transformedNewObj runtime.Object, err error) {
			obj := oldObj.(*Grafana).DeepCopy()
			obj.Spec.Endpoint = "test"
			obj.SetLabels(map[string]string{"key": "value"})
			return obj, nil
		}), nil, nil, false, nil)
		Ω(err).To(Succeed())
		grafana, ok = obj.(*Grafana)
		Ω(ok).To(BeTrue())
		Ω(grafana.Spec.Endpoint).To(Equal("test"))
		objs, err = s.List(ctx, &metainternalversion.ListOptions{LabelSelector: labels.SelectorFromSet(map[string]string{"key": "value"})})
		Ω(err).To(Succeed())
		grafanas, ok = objs.(*GrafanaList)
		Ω(ok).To(BeTrue())
		Ω(len(grafanas.Items)).To(Equal(3))

		By("Test Delete Grafana")
		obj, _, err = s.Delete(ctx, "example2", nil, nil)
		Ω(err).To(Succeed())
		grafana, ok = obj.(*Grafana)
		Ω(ok).To(BeTrue())
		Ω(grafana.Spec.Endpoint).To(Equal("2"))
		objs, err = s.List(ctx, &metainternalversion.ListOptions{LabelSelector: labels.SelectorFromSet(map[string]string{"key": "value"})})
		Ω(err).To(Succeed())
		grafanas, ok = objs.(*GrafanaList)
		Ω(ok).To(BeTrue())
		Ω(len(grafanas.Items)).To(Equal(2))
	})

})
