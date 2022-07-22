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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/utils/pointer"

	"github.com/kubevela/prism/pkg/apis/o11y/config"
	grafanav1alpha1 "github.com/kubevela/prism/pkg/apis/o11y/grafana/v1alpha1"
	"github.com/kubevela/prism/pkg/util/apiserver"
	"github.com/kubevela/prism/pkg/util/subresource"
	_ "github.com/kubevela/prism/test/bootstrap"
	testutil "github.com/kubevela/prism/test/util"
)

func TestGrafanaDatasource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GrafanaDatasource Extension API Test")
}

var _ = Describe("Test GrafanaDatasource API", func() {

	var mockServer *httptest.Server
	var data map[string][]byte

	BeforeEach(func() {
		Ω(testutil.CreateNamespace(config.ObservabilityNamespace)).To(Succeed())
		data = map[string][]byte{}
		mockServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			p := request.Method + " " + request.URL.Path
			switch {
			case p == "POST /api/datasources/":
				bs, _ := io.ReadAll(request.Body)
				uid := apiserver.GetStringFromRawExtension(&runtime.RawExtension{Raw: bs}, "uid")
				data[uid] = bs
				_, _ = writer.Write(bs)
				writer.WriteHeader(http.StatusOK)
			case strings.HasPrefix(p, "PUT /api/datasources/"):
				id := strings.TrimPrefix(p, "PUT /api/datasources/")
				bs, _ := io.ReadAll(request.Body)
				for k, v := range data {
					var m map[string]interface{}
					_ = json.Unmarshal(v, &m)
					if fmt.Sprintf("%d", int(m["id"].(float64))) == id {
						data[k] = bs
						break
					}
				}
				_, _ = writer.Write(bs)
				writer.WriteHeader(http.StatusOK)
			case strings.HasPrefix(p, "GET /api/datasources/uid/"):
				uid := strings.TrimPrefix(p, "GET /api/datasources/uid/")
				db, ok := data[uid]
				if ok {
					_, _ = writer.Write(db)
					writer.WriteHeader(http.StatusOK)
				} else {
					writer.WriteHeader(http.StatusNotFound)
				}
			case strings.HasPrefix(p, "GET /api/datasources"):
				var dbs []string
				for _, val := range data {
					dbs = append(dbs, string(val))
				}
				_, _ = writer.Write([]byte("[" + strings.Join(dbs, ",") + "]"))
				writer.WriteHeader(http.StatusOK)
			case strings.HasPrefix(p, "DELETE /api/datasources/uid/"):
				uid := strings.TrimPrefix(p, "DELETE /api/datasources/uid/")
				if _, ok := data[uid]; ok {
					delete(data, uid)
					writer.WriteHeader(http.StatusOK)
				} else {
					writer.WriteHeader(http.StatusNotFound)
				}
			default:
				writer.WriteHeader(http.StatusNotFound)
			}
		}))
	})

	AfterEach(func() {
		Ω(testutil.DeleteNamespace(config.ObservabilityNamespace)).To(Succeed())
		mockServer.Close()
	})

	It("Test GrafanaDatasource API", func() {
		s := &GrafanaDatasource{}
		By("Test meta info")
		By("Test meta info")
		Ω(s.New()).To(Equal(&GrafanaDatasource{}))
		Ω(s.GetObjectMeta()).To(Equal(&metav1.ObjectMeta{}))
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
				Endpoint: mockServer.URL,
				Access:   grafanav1alpha1.AccessCredential{Token: pointer.String("mock")},
			},
		}
		_, err := (&grafanav1alpha1.Grafana{}).Create(ctx, grafana, nil, nil)
		Ω(err).To(Succeed())

		By("Test Create GrafanaDatasource")
		_, err = s.Create(ctx, &GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "alpha"},
			Spec:       runtime.RawExtension{Raw: []byte(`{"id":0,"key":"val"}`)},
		}, nil, nil)
		Ω(err).To(Succeed())
		_, err = s.Create(ctx, &GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "beta"},
			Spec:       runtime.RawExtension{Raw: []byte(`{"id":1,"key":"value"}`)},
		}, nil, nil)
		Ω(err).To(Succeed())

		By("Test Update GrafanaDatasource")
		_, _, err = s.Update(ctx, "beta", rest.DefaultUpdatedObjectInfo(&GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "beta"},
			Spec:       runtime.RawExtension{Raw: []byte(`{"id":1,"key":"v"}`)},
		}), nil, nil, false, nil)
		Ω(err).To(Succeed())

		By("Test Get GrafanaDatasource")
		obj, err := s.Get(ctx, "alpha", nil)
		Ω(err).To(Succeed())
		gdb, ok := obj.(*GrafanaDatasource)
		Ω(ok).To(BeTrue())
		Ω(gdb.Spec.Raw).To(Equal([]byte(`{"id":0,"key":"val","uid":"alpha"}`)))

		By("Test List GrafanaDatasource")
		objs, err := s.List(ctx, nil)
		Ω(err).To(Succeed())
		dbs, ok := objs.(*GrafanaDatasourceList)
		Ω(ok).To(BeTrue())
		Ω(len(dbs.Items)).To(Equal(2))

		By("Test Delete GrafanaDatasource")
		_, _, err = s.Delete(ctx, "alpha", nil, nil)
		Ω(err).To(Succeed())
		objs, err = s.List(ctx, nil)
		Ω(err).To(Succeed())
		dbs, ok = objs.(*GrafanaDatasourceList)
		Ω(ok).To(BeTrue())
		Ω(len(dbs.Items)).To(Equal(1))

		By("Test GrafanaDatasource Printer")
		_, err = s.ConvertToTable(ctx, obj, nil)
		Ω(err).To(Succeed())
		_, err = s.ConvertToTable(ctx, objs, nil)
		Ω(err).To(Succeed())
	})

})
