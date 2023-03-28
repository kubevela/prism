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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubevela/pkg/util/apiserver"

	"github.com/kubevela/prism/pkg/apis/o11y/config"
)

// GrafanaClient client for operate grafana
// +kubebuilder:object:generate=false
type GrafanaClient interface {
	Get(ctx context.Context, name string) (*Grafana, error)
	List(ctx context.Context, options ...client.ListOption) (*GrafanaList, error)
	Create(ctx context.Context, grafana *Grafana) error
	Update(ctx context.Context, grafana *Grafana) error
	Delete(ctx context.Context, grafana *Grafana) error
}

type grafanaClient struct {
	client.Client
}

// NewGrafanaClient create a client for accessing grafana
func NewGrafanaClient(cli client.Client) GrafanaClient {
	return &grafanaClient{Client: cli}
}

func (c *grafanaClient) get(ctx context.Context, name string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := c.Client.Get(ctx, types.NamespacedName{
		Name:      grafanaSecretNamePrefix + name,
		Namespace: config.ObservabilityNamespace}, secret)
	return secret, err
}

func (c *grafanaClient) Get(ctx context.Context, name string) (*Grafana, error) {
	secret, err := c.get(ctx, name)
	if err != nil {
		return nil, err
	}
	return NewGrafanaFromSecret(secret)
}

func (c *grafanaClient) List(ctx context.Context, options ...client.ListOption) (*GrafanaList, error) {
	opts := apiserver.NewListOptions(options...)
	opts.Namespace = config.ObservabilityNamespace
	secrets := &corev1.SecretList{}
	if err := c.Client.List(ctx, secrets, opts); err != nil {
		return nil, err
	}
	grafanaList := &GrafanaList{}
	for _, secret := range secrets.Items {
		grafana, err := NewGrafanaFromSecret(secret.DeepCopy())
		if err != nil {
			continue
		}
		grafanaList.Items = append(grafanaList.Items, *grafana)
	}
	return grafanaList, nil
}

func (c *grafanaClient) Create(ctx context.Context, grafana *Grafana) error {
	return c.Client.Create(ctx, grafana.ToSecret())
}

func (c *grafanaClient) Update(ctx context.Context, grafana *Grafana) error {
	return c.Client.Update(ctx, grafana.ToSecret())
}

func (c *grafanaClient) Delete(ctx context.Context, grafana *Grafana) error {
	return c.Client.Delete(ctx, grafana.ToSecret())
}
