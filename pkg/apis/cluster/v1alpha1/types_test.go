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

	clustergatewayv1alpha1 "github.com/oam-dev/cluster-gateway/pkg/apis/cluster/v1alpha1"
	clustergatewaycommon "github.com/oam-dev/cluster-gateway/pkg/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ocmclusterv1 "open-cluster-management.io/api/cluster/v1"

	"github.com/kubevela/prism/pkg/util/singleton"
)

var _ = Describe("Test Cluster API", func() {

	It("Test Cluster API", func() {
		c := &Cluster{}
		c.SetName("example")
		Ω(c.GetFullName()).To(Equal("example"))
		c.Spec.Alias = "alias"
		Ω(c.GetFullName()).To(Equal("example (alias)"))

		By("Test meta info")
		Ω(c.New()).To(Equal(&Cluster{}))
		Ω(c.NamespaceScoped()).To(BeFalse())
		Ω(c.ShortNames()).To(SatisfyAll(
			ContainElement("vc"),
			ContainElement("vela-cluster"),
			ContainElement("vela-clusters"),
		))
		Ω(c.GetGroupVersionResource().GroupVersion()).To(Equal(GroupVersion))
		Ω(c.GetGroupVersionResource().Resource).To(Equal(ClusterResource))
		Ω(c.IsStorageVersion()).To(BeTrue())
		Ω(c.NewList()).To(Equal(&ClusterList{}))

		ctx := context.Background()

		By("Create storage namespace")
		StorageNamespace = "vela-system"
		Ω(singleton.GetKubeClient().Create(ctx, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: StorageNamespace}})).To(Succeed())

		By("Create cluster secret")
		Ω(singleton.GetKubeClient().Create(ctx, &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: StorageNamespace,
				Labels: map[string]string{
					clustergatewaycommon.LabelKeyClusterCredentialType: string(clustergatewayv1alpha1.CredentialTypeX509Certificate),
					clustergatewaycommon.LabelKeyClusterEndpointType:   string(clustergatewayv1alpha1.ClusterEndpointTypeConst),
					"key": "value",
				},
			},
		})).To(Succeed())
		Ω(singleton.GetKubeClient().Create(ctx, &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-invalid",
				Namespace: StorageNamespace,
			},
		})).To(Succeed())

		By("Test get cluster from cluster secret")
		obj, err := c.Get(ctx, "test-cluster", nil)
		Ω(err).To(Succeed())
		cluster, ok := obj.(*Cluster)
		Ω(ok).To(BeTrue())
		Ω(cluster.Spec.CredentialType).To(Equal(clustergatewayv1alpha1.CredentialTypeX509Certificate))
		Ω(cluster.GetLabels()["key"]).To(Equal("value"))

		_, err = c.Get(ctx, "cluster-invalid", nil)
		Ω(err).To(Satisfy(IsInvalidClusterSecretError))

		By("Create OCM ManagedCluster")
		Ω(singleton.GetKubeClient().Create(ctx, &ocmclusterv1.ManagedCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ocm-bad-cluster",
				Namespace: StorageNamespace,
			},
		})).To(Succeed())
		Ω(singleton.GetKubeClient().Create(ctx, &ocmclusterv1.ManagedCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ocm-cluster",
				Namespace: StorageNamespace,
				Labels:    map[string]string{"key": "value"},
			},
			Spec: ocmclusterv1.ManagedClusterSpec{
				ManagedClusterClientConfigs: []ocmclusterv1.ClientConfig{{URL: "test-url"}},
			},
		})).To(Succeed())
		Ω(singleton.GetKubeClient().Create(ctx, &ocmclusterv1.ManagedCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: StorageNamespace,
				Labels:    map[string]string{"key": "value-dup"},
			},
			Spec: ocmclusterv1.ManagedClusterSpec{
				ManagedClusterClientConfigs: []ocmclusterv1.ClientConfig{{URL: "test-url-dup"}},
			},
		})).To(Succeed())

		By("Test get cluster from OCM managed cluster")
		_, err = c.Get(ctx, "ocm-bad-cluster", nil)
		Ω(err).To(Satisfy(IsInvalidManagedClusterError))

		obj, err = c.Get(ctx, "ocm-cluster", nil)
		Ω(err).To(Succeed())
		cluster, ok = obj.(*Cluster)
		Ω(ok).To(BeTrue())
		Expect(cluster.Spec.CredentialType).To(Equal(CredentialTypeOCMManagedCluster))

		By("Test get local cluster")
		obj, err = c.Get(ctx, "local", nil)
		Ω(err).To(Succeed())
		cluster, ok = obj.(*Cluster)
		Ω(ok).To(BeTrue())
		Expect(cluster.Spec.CredentialType).To(Equal(CredentialTypeInternal))

		_, err = c.Get(ctx, "cluster-not-exist", nil)
		Ω(err).To(Satisfy(apierrors.IsNotFound))

		By("Test list clusters")
		objs, err := c.List(ctx, nil)
		Ω(err).To(Succeed())
		clusters, ok := objs.(*ClusterList)
		Ω(ok).To(BeTrue())
		Expect(len(clusters.Items)).To(Equal(3))

		objs, err = c.List(ctx, &metainternalversion.ListOptions{LabelSelector: labels.SelectorFromSet(map[string]string{"key": "value"})})
		Ω(err).To(Succeed())
		clusters, ok = objs.(*ClusterList)
		Ω(ok).To(BeTrue())
		Expect(len(clusters.Items)).To(Equal(2))

		By("Test print table")
		_, err = c.ConvertToTable(ctx, cluster, nil)
		Ω(err).To(Succeed())
		_, err = c.ConvertToTable(ctx, clusters, nil)
		Ω(err).To(Succeed())
	})

})
