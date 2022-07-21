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
	"fmt"
	"net/http"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	grafanav1alpha1 "github.com/kubevela/prism/pkg/apis/o11y/grafana/v1alpha1"
	"github.com/kubevela/prism/pkg/util/subresource"
)

// GrafanaDatasourceClient client for grafana datasource
// +kubebuilder:object:generate=false
type GrafanaDatasourceClient interface {
	Get(ctx context.Context, name string) (*GrafanaDatasource, error)
	//List(ctx context.Context, options ...client.ListOption) (*GrafanaDatasourceList, error)
	Create(ctx context.Context, grafanaDatasource *GrafanaDatasource) error
	Update(ctx context.Context, grafanaDatasource *GrafanaDatasource) error
	Delete(ctx context.Context, grafanaDatasource *GrafanaDatasource) error
}

// NewGrafanaDatasourceClient create GrafanaDatasourceClient
func NewGrafanaDatasourceClient(cli client.Client) GrafanaDatasourceClient {
	return &grafanaDatasourceClient{
		GrafanaClient: grafanav1alpha1.NewGrafanaClient(cli),
	}
}

type grafanaDatasourceClient struct {
	grafanav1alpha1.GrafanaClient
}

func (in *grafanaDatasourceClient) Get(ctx context.Context, name string) (*GrafanaDatasource, error) {
	resourceName := subresource.NewCompoundName(name)
	datasource := &GrafanaDatasource{
		ObjectMeta: metav1.ObjectMeta{Name: resourceName.String(), UID: "-"},
	}
	return datasource, grafanav1alpha1.NewGrafanaSubResourceRequest(datasource, name).
		WithMethod(http.MethodGet).
		WithPathFunc(func() (string, error) {
			return "/api/datasources/name/" + url.PathEscape(resourceName.SubResourceName), nil
		}).
		WithOnSuccess(func(respBody []byte) error {
			datasource.UID = ""
			datasource.Spec = runtime.RawExtension{Raw: respBody}
			return nil
		}).
		Do(ctx, in.GrafanaClient)
}

func (in *grafanaDatasourceClient) Create(ctx context.Context, datasource *GrafanaDatasource) error {
	return grafanav1alpha1.NewGrafanaSubResourceRequest(datasource, datasource.GetName()).
		WithMethod(http.MethodPost).
		WithPathFunc(func() (string, error) {
			return "/api/datasources/", nil
		}).
		WithBodyFunc(func() ([]byte, error) {
			return datasource.Spec.Raw, nil
		}).
		WithOnSuccess(func(respBody []byte) error {
			return datasource.FromGrafanaAPIResponse(respBody)
		}).
		Do(ctx, in.GrafanaClient)
}

func (in *grafanaDatasourceClient) Update(ctx context.Context, datasource *GrafanaDatasource) error {
	return grafanav1alpha1.NewGrafanaSubResourceRequest(datasource, datasource.GetName()).
		WithMethod(http.MethodPut).
		WithPathFunc(func() (string, error) {
			id, err := datasource.GetGrafanaDatasourceID()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("/api/datasources/%d", id), nil
		}).
		WithBodyFunc(func() ([]byte, error) {
			return datasource.Spec.Raw, nil
		}).
		WithOnSuccess(func(respBody []byte) error {
			return datasource.FromGrafanaAPIResponse(respBody)
		}).
		Do(ctx, in.GrafanaClient)
}

func (in *grafanaDatasourceClient) Delete(ctx context.Context, datasource *GrafanaDatasource) error {
	return grafanav1alpha1.NewGrafanaSubResourceRequest(datasource, datasource.GetName()).
		WithMethod(http.MethodDelete).
		WithPathFunc(func() (string, error) {
			return "/api/datasources/name/" + subresource.NewCompoundName(datasource.GetName()).SubResourceName, nil
		}).
		Do(ctx, in.GrafanaClient)
}
