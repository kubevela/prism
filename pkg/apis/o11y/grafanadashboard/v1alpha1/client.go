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
	"net/http"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	grafanav1alpha1 "github.com/kubevela/prism/pkg/apis/o11y/grafana/v1alpha1"
	"github.com/kubevela/prism/pkg/util/subresource"
)

// GrafanaDashboardClient client for grafana datasource
// +kubebuilder:object:generate=false
type GrafanaDashboardClient interface {
	Get(ctx context.Context, name string) (*GrafanaDashboard, error)
	//List(ctx context.Context, options ...client.ListOption) (*GrafanaDashboardList, error)
	Create(ctx context.Context, GrafanaDashboard *GrafanaDashboard) error
	Update(ctx context.Context, GrafanaDashboard *GrafanaDashboard) error
	Delete(ctx context.Context, GrafanaDashboard *GrafanaDashboard) error
}

// NewGrafanaDashboardClient create GrafanaDashboardClient
func NewGrafanaDashboardClient(cli client.Client) GrafanaDashboardClient {
	return &grafanaDashboardClient{
		GrafanaClient: grafanav1alpha1.NewGrafanaClient(cli),
	}
}

type grafanaDashboardClient struct {
	grafanav1alpha1.GrafanaClient
}

func (in *grafanaDashboardClient) Get(ctx context.Context, name string) (*GrafanaDashboard, error) {
	resourceName := subresource.NewCompoundName(name)
	dashboard := &GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{Name: resourceName.String(), UID: "-"},
	}
	return dashboard, grafanav1alpha1.NewGrafanaSubResourceRequest(dashboard, name).
		WithMethod(http.MethodGet).
		WithPathFunc(func() (string, error) {
			return "/api/dashboards/uid/" + url.PathEscape(resourceName.SubResourceName), nil
		}).
		WithOnSuccess(dashboard.FromBody).
		Do(ctx, in.GrafanaClient)
}

func (in *grafanaDashboardClient) Create(ctx context.Context, dashboard *GrafanaDashboard) error {
	return grafanav1alpha1.NewGrafanaSubResourceRequest(dashboard, dashboard.GetName()).
		WithMethod(http.MethodPost).
		WithPathFunc(func() (string, error) {
			return "/api/dashboards/db", nil
		}).
		WithBodyFunc(dashboard.ToBody).
		Do(ctx, in.GrafanaClient)
}

func (in *grafanaDashboardClient) Update(ctx context.Context, dashboard *GrafanaDashboard) error {
	return in.Create(ctx, dashboard)
}

func (in *grafanaDashboardClient) Delete(ctx context.Context, dashboard *GrafanaDashboard) error {
	resourceName := subresource.NewCompoundName(dashboard.GetName())
	return grafanav1alpha1.NewGrafanaSubResourceRequest(dashboard, dashboard.GetName()).
		WithMethod(http.MethodDelete).
		WithPathFunc(func() (string, error) {
			return "/api/dashboards/uid/" + url.PathEscape(resourceName.SubResourceName), nil
		}).
		Do(ctx, in.GrafanaClient)
}
