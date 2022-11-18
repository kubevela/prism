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

package apiserver_test

import (
	"context"
	"fmt"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	"k8s.io/apiserver/pkg/server"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"

	apiserver "github.com/kubevela/prism/pkg/dynamicapiserver"
	"github.com/kubevela/prism/pkg/util/singleton"
	_ "github.com/kubevela/prism/test/bootstrap"
	"github.com/kubevela/prism/test/util"
)

func TestDynamicServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Dynamic Server")
}

func createConfigMapForDiscovery(group, version, kind string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	resource := strings.ToLower(kind) + "s"
	cm.SetName(fmt.Sprintf("%s.%s.%s", resource, group, version))
	cm.SetNamespace("vela-system")
	cm.Data = map[string]string{}
	cm.Data["encoder"] = fmt.Sprintf(`
	    parameter: {
	        apiVersion: "%s/%s"
	        kind: "%s"
	        metadata: {...}
	    }
	    output: {
	        apiVersion: "v1"
	        kind: "ConfigMap"
	        metadata: parameter.metadata
			data: {}
	    }
	`, group, version, kind)
	cm.Data["decoder"] = fmt.Sprintf(`
	    parameter: {
	        apiVersion: "v1"
	        kind: "ConfigMap"
	        metadata: {...}
			data: {...}
	    }
	    output: {
	        apiVersion: "%s/%s"
	        kind: "%s"
	        metadata: parameter.metadata
	    }
	`, group, version, kind)
	return cm
}

var _ = Describe("Test dynamic server", func() {
	It("Test bootstrap and mutate spec", func() {
		By("Bootstrap")
		_ = util.CreateNamespace("vela-system")
		s := &builder.GenericAPIServer{
			Handler: &server.APIServerHandler{
				GoRestfulContainer: restful.NewContainer(),
			},
			DiscoveryGroupManager:      discovery.NewRootAPIsHandler(nil, nil),
			EquivalentResourceRegistry: runtime.NewEquivalentResourceRegistry(),
		}
		singleton.InitGenericAPIServer(s)
		cfg := &server.RecommendedConfig{}
		singleton.InitServerConfig(cfg)
		stopCh := make(chan struct{})
		defer close(stopCh)
		_ = apiserver.StartDefaultDynamicAPIServer(server.PostStartHookContext{StopCh: stopCh})

		By("Add Resource API")
		ctx := context.Background()
		cm1 := createConfigMapForDiscovery("test.oam.dev", "v1alpha1", "Test")
		Ω(singleton.KubeClient.Get().Create(ctx, cm1)).To(Succeed())
		Eventually(func(g Gomega) {
			g.Ω(slices.IndexFunc(s.Handler.GoRestfulContainer.RegisteredWebServices(), func(_ws *restful.WebService) bool {
				return _ws.RootPath() == path.Join(server.APIGroupPrefix, "test.oam.dev", "v1alpha1")
			}) >= 0).Should(BeTrue())
		}).WithTimeout(5 * time.Second).WithPolling(1 * time.Second).Should(Succeed())

		By("Update & Delete Resource API")
		Ω(singleton.KubeClient.Get().Delete(ctx, cm1)).To(Succeed())
		cm2 := createConfigMapForDiscovery("next.test.oam.dev", "v1alpha1", "Test")
		Ω(singleton.KubeClient.Get().Create(ctx, cm2)).To(Succeed())
		cm2.Data = createConfigMapForDiscovery("new.test.oam.dev", "v1alpha1", "Test").Data
		Ω(singleton.KubeClient.Get().Update(ctx, cm2)).To(Succeed())
		cm3 := createConfigMapForDiscovery("next.new.test.oam.dev", "v1alpha1", "Test")
		Ω(singleton.KubeClient.Get().Create(ctx, cm3)).To(Succeed())
		Ω(singleton.KubeClient.Get().Delete(ctx, cm3)).To(Succeed())
		Eventually(func(g Gomega) {
			webservices := s.Handler.GoRestfulContainer.RegisteredWebServices()
			g.Ω(slices.IndexFunc(webservices, func(_ws *restful.WebService) bool {
				return _ws.RootPath() == path.Join(server.APIGroupPrefix, "test.oam.dev", "v1alpha1")
			}) < 0).Should(BeTrue())
			g.Ω(slices.IndexFunc(webservices, func(_ws *restful.WebService) bool {
				return _ws.RootPath() == path.Join(server.APIGroupPrefix, "next.test.oam.dev", "v1alpha1")
			}) < 0).Should(BeTrue())
			g.Ω(slices.IndexFunc(webservices, func(_ws *restful.WebService) bool {
				return _ws.RootPath() == path.Join(server.APIGroupPrefix, "new.test.oam.dev", "v1alpha1")
			}) >= 0).Should(BeTrue())
			g.Ω(slices.IndexFunc(webservices, func(_ws *restful.WebService) bool {
				return _ws.RootPath() == path.Join(server.APIGroupPrefix, "new.next.test.oam.dev", "v1alpha1")
			}) < 0).Should(BeTrue())
		}).WithTimeout(15 * time.Second).WithPolling(1 * time.Second).Should(Succeed())
	})
})
