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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/kubevela/prism/pkg/apis/o11y/config"
	grafanav1alpha1 "github.com/kubevela/prism/pkg/apis/o11y/grafana/v1alpha1"
	"github.com/kubevela/prism/pkg/util/subresource"
	testutil "github.com/kubevela/prism/test/util"
)

var _ = Describe("Test GrafanaDatasource API", func() {

	BeforeEach(func() {
		Ω(testutil.CreateNamespace(config.ObservabilityNamespace)).To(Succeed())
	})

	AfterEach(func() {
		Ω(testutil.DeleteNamespace(config.ObservabilityNamespace)).To(Succeed())
	})

	It("Test Grafana API", func() {
		s := &GrafanaDatasource{}
		By("Test meta info")
		By("Test meta info")
		Ω(s.New()).To(Equal(&GrafanaDatasource{}))
		Ω(s.NamespaceScoped()).To(BeFalse())
		Ω(s.ShortNames()).To(ContainElement("gds"))
		Ω(s.GetGroupVersionResource().GroupVersion()).To(Equal(GroupVersion))
		Ω(s.GetGroupVersionResource().Resource).To(Equal(GrafanaDatasourceResource))
		Ω(s.IsStorageVersion()).To(BeTrue())
		Ω(s.NewList()).To(Equal(&GrafanaDatasourceList{}))

		ctx := context.Background()

		By("Create Grafana")
		grafana := &grafanav1alpha1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Name: subresource.DefaultParentResourceName},
			Spec: grafanav1alpha1.GrafanaSpec{
				Endpoint: "mock",
				Access:   grafanav1alpha1.AccessCredential{Token: pointer.String("mock")},
			},
		}
		_, err := (&grafanav1alpha1.Grafana{}).Create(ctx, grafana, nil, nil)
		Ω(err).To(Succeed())

	})

})
